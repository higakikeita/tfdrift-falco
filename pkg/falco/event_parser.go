package falco

import (
	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// parseFalcoOutput parses a Falco output response into a TFDrift event
// Supports AWS CloudTrail, GCP Audit Log, and Azure Audit Log events
func (s *Subscriber) parseFalcoOutput(res *outputs.Response) *types.Event {
	// Handle nil response
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	switch res.Source {
	case "aws_cloudtrail":
		return s.parseAWSEvent(res)
	case "gcpaudit":
		return s.gcpParser.Parse(res)
	case "azureaudit":
		return s.azureParser.Parse(res)
	default:
		log.Debugf("Unknown Falco source: %s", res.Source)
		return nil
	}
}
