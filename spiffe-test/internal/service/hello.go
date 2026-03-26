package service

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

type helloService struct {
	allowedIDs map[string]struct{}
}

func NewHelloService(allowedSPIFFEIDs []string) HelloService {
	allowed := make(map[string]struct{}, len(allowedSPIFFEIDs))
	for _, id := range allowedSPIFFEIDs {
		allowed[id] = struct{}{}
	}
	return &helloService{
		allowedIDs: allowed,
	}
}

func (s *helloService) SayHello(ctx context.Context, callerID spiffeid.ID) (string, error) {
	id := callerID.String()
	if _, ok := s.allowedIDs[id]; !ok {
		return "", fmt.Errorf("%w: SPIFFE ID %s is not in the allowed list", ErrUnauthorized, id)
	}
	return fmt.Sprintf("Hello, %s!", id), nil
}
