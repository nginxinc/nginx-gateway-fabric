package policies

import (
	"regexp"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

func ValidateSize(s *ngfAPI.Size) error {
	if s == nil {
		return nil
	}

	_, err := regexp.MatchString(`^\d{1,4}(m|g|k)+$`, string(*s))

	return err
}

func ValidateDuration(d *ngfAPI.Duration) error {
	if d == nil {
		return nil
	}

	_, err := regexp.MatchString(`^\d{1,4}(ms|s)?$`, string(*d))

	return err
}
