package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	DesiredImage string
	AppLabel     string
	Timeout      time.Duration
	DrainTimeout time.Duration
	KubectlPath  string
	Ignore       []string
}

const (
	RackLabel string = "cke.cybozu.com/rack"
)

var (
	cmd = cobra.Command{
		Use:     "daemonset-updater",
		Short:   "daemnonset-updater updates daemonsets that is on-delete strategy",
		Run:     cmdMain,
		Version: "1.0.0",
	}

	cfg Config

	log *slog.Logger
)

func init() {
	cmd.Flags().StringVarP(&cfg.DesiredImage, "desired-image", "d", "", "Desired image")
	cmd.Flags().StringVarP(&cfg.AppLabel, "app-label", "l", "", "The label associated with pods that is a part of the daemonset. exp) app=test")
	cmd.Flags().DurationVar(&cfg.Timeout, "timeout", 9*time.Hour, "Total timeout to update")
	cmd.Flags().DurationVar(&cfg.DrainTimeout, "drain-timeout", 30*time.Minute, "Timeout for draining")
	cmd.Flags().StringVar(&cfg.KubectlPath, "kubectl-path", "/usr/bin/kubectl", "Path to kubectl")
	cmd.Flags().StringSliceVar(&cfg.Ignore, "ignore", []string{}, "List of nodes that is ignored")

	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func cmdMain(cmd *cobra.Command, args []string) {

	if cfg.DesiredImage == "" {
		log.Error("desired image is not set")
		return
	}

	if cfg.AppLabel == "" {
		log.Error("app label is not set")
		return
	}

	ticker := time.NewTicker(cfg.Timeout)

	appLabelKeyValue := strings.Split(cfg.AppLabel, "=")
	if len(appLabelKeyValue) != 2 {
		log.Error("app label is invalid", slog.String("app-label", cfg.AppLabel))
		return
	}
	appLabelKey := appLabelKeyValue[0]
	appLabelValue := appLabelKeyValue[1]

	config, err := config.GetConfig()
	if err != nil {
		log.Error("failed to get config", slog.Any("error", err))
		return
	}

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		log.Error("failed to create k8s client", slog.Any("error", err))
		return
	}

	ctx := context.Background()

	nodeList := corev1.NodeList{}
	if err := c.List(ctx, &nodeList); err != nil {
		log.ErrorContext(ctx, "failed to list nodes", slog.Any("error", err))
		return
	}

	// get node lists per a rack
	rackList := make(map[int][]string)
	for _, node := range nodeList.Items {
		// ignore the node whether it is in the ignore list
		if slices.Contains(cfg.Ignore, node.GetName()) {
			log.InfoContext(ctx, "ignore to update", slog.String("node", node.GetName()))
			continue
		}
		rackNumber, ok := node.Labels[RackLabel]
		if !ok {
			log.WarnContext(ctx, "the rack label is not found", slog.String("node", node.GetName()))
			continue
		}
		n, err := strconv.Atoi(rackNumber)
		if err != nil {
			log.WarnContext(ctx, "the value of the rack label is invalid", slog.String("node", node.GetName()), slog.String("value", rackNumber))
			continue
		}
		if _, ok = rackList[n]; !ok {
			rackList[n] = []string{node.GetName()}
		} else {
			rackList[n] = append(rackList[n], node.GetName())
		}
	}

	// get a pod list of the daemonset
	podList := corev1.PodList{}
	listOption := client.MatchingLabels{
		appLabelKey: appLabelValue,
	}
	if err := c.List(ctx, &podList, listOption); err != nil {
		log.ErrorContext(ctx, "failed to list pods", slog.Any("error", err))
		return
	}

	podNodeMap := make(map[string]*corev1.Pod)
	for _, pod := range podList.Items {
		podNodeMap[pod.Spec.NodeName] = &pod
	}

	// get nodes that is not desired image per a rack
	updateList := make(map[int][]string)
	for rackNumber, nodeList := range rackList {
		updateNodeList := make([]string, 0, len(nodeList))
		for _, node := range nodeList {
			pod, ok := podNodeMap[node]
			if !ok {
				log.WarnContext(ctx, "failed to find the corresponding pod to the node", slog.String("node", node))
				continue
			}

			// check the image is the desired one.
			// should we use .spec.containers[0]?
			image := pod.Spec.Containers[0].Image
			if image != cfg.DesiredImage {
				log.InfoContext(ctx, "image is not desired", slog.String("node", node), slog.String("pod", pod.GetName()))
				updateNodeList = append(updateNodeList, node)
			}
		}
		if len(updateNodeList) > 0 {
			updateList[rackNumber] = updateNodeList
		}
		log.InfoContext(ctx, "the rack has nodes that is needed to update", slog.Int("rack", rackNumber), slog.Any("nodes", updateNodeList))
	}

	ctx, cancel := context.WithCancel(ctx)
	// var wg sync.WaitGroup

	targetChan := make(chan string)
	complete := make(chan struct{})
	go func() {
		for rackNumber, nodeList := range updateList {
			log.InfoContext(ctx, "start to update", slog.Int("rack", rackNumber))

			for _, node := range nodeList {
				targetChan <- node
				// drain(dry-run)
				log.InfoContext(ctx, "start to drain", slog.String("node", node), slog.Bool("dry-run", true))
				if err := drain(ctx, node, cfg.DrainTimeout, true); err != nil {
					log.WarnContext(ctx, "failed to drain", slog.String("node", node), slog.Bool("dry-run", true), slog.Any("error", err))
					// if we fail to drain, even if it is dry-run, go next.
					continue
				}
				// drain
				log.InfoContext(ctx, "start to drain", slog.String("node", node), slog.Bool("dry-run", false))
				if err := drain(ctx, node, cfg.DrainTimeout, false); err != nil {
					log.WarnContext(ctx, "failed to drain", slog.String("node", node), slog.Bool("dry-run", false), slog.Any("error", err))
					// if we fail to drain, go next.
					continue
				}
				// delete the pod
				pod := podNodeMap[node]
				log.InfoContext(ctx, "delete the target pod", slog.String("node", node), slog.String("pod", pod.GetName()))
				if err := c.Delete(ctx, pod); err != nil {
					log.WarnContext(ctx, "failed to delete pod", slog.String("node", node), slog.String("name", pod.GetName()), slog.Any("error", err))
					continue
				}
				// uncordon
				log.InfoContext(ctx, "start to uncordon", slog.String("node", node))
				if err := uncordon(ctx, node); err != nil {
					log.WarnContext(ctx, "failed to uncordon", slog.String("node", node))
					// Can we continue to update?
					// I'm not sure that it is ok to leave the node that is SchedulingDisabled
					continue
				}
			}
		}

		// complete
		complete <- struct{}{}
	}()

	var target string
	for {
		select {
		case <-ticker.C:
			// time is up
			// cancel the running procedure
			cancel()
			// uncordon anyway(can we do?)
			if err := uncordon(context.Background(), target); err != nil {
				log.ErrorContext(ctx, "failed to uncordon after canceled", slog.String("node", target))
				return
			}
			return
		case t := <-targetChan:
			target = t
			log.InfoContext(ctx, "start to handle", slog.String("node", t))
		case <-complete:
			log.InfoContext(ctx, "complete to update")
			// it's ok to reach here without calling cancel()
			return
		}
	}
}

func drain(ctx context.Context, node string, timeout time.Duration, dryRun bool) error {
	args := []string{"drain", node, "--delete-emptydir-data=true", "--ignore-daemonsets=true", fmt.Sprintf("--timeout=%s", timeout)}
	if dryRun {
		args = append(args, "--dry-run=server")
	}
	stdout, stderr, err := kubectl(ctx, args...)
	if err != nil {
		return fmt.Errorf("stdout: %s, stderr: %s", stdout, stderr)
	}
	return nil
}

func uncordon(ctx context.Context, node string) error {
	_, _, err := kubectl(ctx, "uncordon", node)
	if err != nil {
		return err
	}
	return nil
}

func kubectl(ctx context.Context, args ...string) ([]byte, []byte, error) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	c := exec.CommandContext(ctx, cfg.KubectlPath, args...)
	c.Stdout = &outBuf
	c.Stderr = &errBuf

	if err := c.Start(); err != nil {
		return outBuf.Bytes(), errBuf.Bytes(), err
	}

	if err := c.Wait(); err != nil {
		return outBuf.Bytes(), errBuf.Bytes(), err
	}
	return outBuf.Bytes(), errBuf.Bytes(), nil
}
