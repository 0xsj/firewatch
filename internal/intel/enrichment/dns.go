package enrichment

import (
	"context"
	"fmt"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/netutil"
)

// DNS performs reverse DNS lookups on IP-type IOCs.
type DNS struct{}

func NewDNS() *DNS { return &DNS{} }

func (d *DNS) Name() string { return "dns" }

func (d *DNS) Enrich(ctx context.Context, ioc *models.IOC) error {
	if ioc.Type != models.IOCTypeIP {
		return nil
	}

	names, err := netutil.ReverseLookupContext(ctx, ioc.Value)
	if err != nil {
		return fmt.Errorf("reverse lookup %s: %w", ioc.Value, err)
	}

	if len(names) > 0 {
		ioc.Hostname = names[0]
	}
	return nil
}
