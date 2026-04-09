package bucketprovisioner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-logr/logr"
	bktv1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	apibkt "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
)

const (
	dummyRegion = "dummy-region"
)

type s3Provisioner struct {
	s3Client        *s3.Client
	endpoint        Endpoint
	accessKeyID     string
	secretAccessKey string
	requestTimeout  time.Duration
	logger          logr.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

var _ apibkt.Provisioner = (*s3Provisioner)(nil)

// New creates a new S3 provisioner from a validated Config.
func New(ctx context.Context, cfg Config, logger logr.Logger) (*s3Provisioner, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(dummyRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			cfg.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint.URL)
		options.UsePathStyle = true
	})

	cctx, cancel := context.WithCancel(ctx)

	return &s3Provisioner{
		s3Client:        s3Client,
		endpoint:        cfg.Endpoint,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
		requestTimeout:  cfg.RequestTimeout,
		logger:          logger,
		ctx:             cctx,
		cancel:          cancel,
	}, nil
}

// Stop cancels the provisioner-scoped context, interrupting any in-flight requests.
func (p *s3Provisioner) Stop() {
	p.cancel()
}

func (p *s3Provisioner) GenerateUserID(obc *bktv1alpha1.ObjectBucketClaim, ob *bktv1alpha1.ObjectBucket) (string, error) {
	return fmt.Sprintf("%s/%s", obc.Namespace, obc.Name), nil
}

func (p *s3Provisioner) Provision(options *apibkt.BucketOptions) (*bktv1alpha1.ObjectBucket, error) {
	ctx, cancel := context.WithTimeout(p.ctx, p.requestTimeout)
	defer cancel()

	p.logger.Info("provisioning bucket", "bucket", options.BucketName)
	if err := p.ensureBucket(ctx, options.BucketName); err != nil {
		return nil, err
	}
	p.logger.Info("bucket provisioned", "bucket", options.BucketName)
	return p.objectBucket(options.BucketName), nil
}

func (p *s3Provisioner) Grant(options *apibkt.BucketOptions) (*bktv1alpha1.ObjectBucket, error) {
	return nil, errors.New("not implemented")
}

func (p *s3Provisioner) Delete(ob *bktv1alpha1.ObjectBucket) error {
	ctx, cancel := context.WithTimeout(p.ctx, p.requestTimeout)
	defer cancel()

	bucket := ob.Spec.Endpoint.BucketName
	p.logger.Info("deleting bucket", "bucket", bucket)
	if err := p.deleteAllObjects(ctx, bucket); err != nil {
		return err
	}

	_, err := p.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		p.logger.Info("bucket deleted", "bucket", bucket)
		return nil
	}

	var noSuchBucket *s3types.NoSuchBucket
	if errors.As(err, &noSuchBucket) {
		return nil
	}
	return err
}

func (p *s3Provisioner) Revoke(ob *bktv1alpha1.ObjectBucket) error {
	return errors.New("not implemented")
}

func (p *s3Provisioner) ensureBucket(ctx context.Context, bucketName string) error {
	_, headErr := p.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucketName)})
	if headErr == nil {
		return nil
	}

	_, err := p.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucketName)})
	if err == nil {
		return nil
	}

	var alreadyOwned *s3types.BucketAlreadyOwnedByYou
	if errors.As(err, &alreadyOwned) {
		return nil
	}
	return err
}

func (p *s3Provisioner) deleteAllObjects(ctx context.Context, bucket string) error {
	paginator := s3.NewListObjectsV2Paginator(p.s3Client, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			var noSuchBucket *s3types.NoSuchBucket
			if errors.As(err, &noSuchBucket) {
				return nil
			}
			return err
		}
		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]s3types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			objects = append(objects, s3types.ObjectIdentifier{Key: obj.Key})
		}

		p.logger.Info("deleting objects", "bucket", bucket, "objects", len(objects))
		_, err = p.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *s3Provisioner) objectBucket(bucketName string) *bktv1alpha1.ObjectBucket {
	return &bktv1alpha1.ObjectBucket{
		Spec: bktv1alpha1.ObjectBucketSpec{
			Connection: &bktv1alpha1.Connection{
				Endpoint: &bktv1alpha1.Endpoint{
					BucketHost:           p.endpoint.Host,
					BucketPort:           p.endpoint.Port,
					BucketName:           bucketName,
					Region:               dummyRegion,
					AdditionalConfigData: map[string]string{},
				},
				Authentication: &bktv1alpha1.Authentication{
					AccessKeys: &bktv1alpha1.AccessKeys{
						AccessKeyID:     p.accessKeyID,
						SecretAccessKey: p.secretAccessKey,
					},
				},
				AdditionalState: map[string]string{},
			},
		},
	}
}
