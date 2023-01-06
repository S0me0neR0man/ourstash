package token

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Tokens struct {
	// Cached accessToken to avoid an STS token exchange for every call to
	// GetRequestMetadata.
	mu            sync.Mutex
	tokenMetadata map[string]string
	tokenExpiry   time.Time
}

func (t *Tokens) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// todo: credentials.CheckSecurityLevel

	// Holding the lock for the whole duration of the STS request and response
	// processing ensures that concurrent RPCs don't end up in multiple
	// requests being made.
	//t.mu.Lock()
	//defer t.mu.Unlock()
	//
	//if md := t.cachedMetadata(); md != nil {
	//	return md, nil
	//}
	//req, err := constructRequest(ctx, t.opts)
	//if err != nil {
	//	return nil, err
	//}
	//respBody, err := sendRequest(t.client, req)
	//if err != nil {
	//	return nil, err
	//}
	//ti, err := tokenInfoFromResponse(respBody)
	//if err != nil {
	//	return nil, err
	//}
	//t.tokenMetadata = map[string]string{"Authorization": fmt.Sprintf("%s %s", ti.tokenType, ti.token)}
	//t.tokenExpiry = ti.expiryTime
	return map[string]string{"Authorization": fmt.Sprintf("token cheto")}, nil
}

func (t *Tokens) RequireTransportSecurity() bool {
	return false
}
