package export

import (
	"github.com/0xsj/firewatch/internal/storage/models"
)

// Exporter converts internal data to an external format.
type Exporter interface {
	// Name returns the format identifier (e.g. "stix", "misp", "csv").
	Name() string

	// ContentType returns the MIME type for the output.
	ContentType() string

	// ExportIOCs serializes IOCs to the target format.
	ExportIOCs(iocs []*models.IOC) ([]byte, error)

	// ExportCampaigns serializes campaigns to the target format.
	ExportCampaigns(campaigns []*models.Campaign) ([]byte, error)
}
