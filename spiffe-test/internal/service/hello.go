package service

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

type helloService struct {
	allowedIDs map[spiffeid.ID]struct{}
}

func NewHelloService(allowedIDs []spiffeid.ID) HelloService {
	allowed := make(map[spiffeid.ID]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		allowed[id] = struct{}{}
	}
	return &helloService{
		allowedIDs: allowed,
	}
}

func (s *helloService) SayHello(ctx context.Context, callerID spiffeid.ID) (string, error) {
	if _, ok := s.allowedIDs[callerID]; !ok {
		return "", fmt.Errorf("%w: SPIFFE ID %s is not in the allowed list", ErrUnauthorized, callerID)
	}
	return fmt.Sprintf("Hello, %s!", callerID), nil
}
