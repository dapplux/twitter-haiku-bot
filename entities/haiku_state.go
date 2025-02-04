package entities

import (
	"database/sql/driver"
	"fmt"
)

type HaikuState string

const (
	HaikuStateCreated          = "created"
	HaikuStateSummaryGetting   = "summary_getting"
	HaikuStateSummaryGot       = "summary_got"
	HaikuStateHaikuTextGetting = "haiku_text_getting"
	HaikuStateHaikuTextGot     = "haiku_text_got"
	HaikuStateComenting        = "comenting"
	HaikuStateDone             = "done"
	HaikuStateFailed           = "failed"
)

// Scan for HaikuState
func (hs *HaikuState) Scan(value interface{}) error {
	if value == nil {
		*hs = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan HaikuState: invalid type %T", value)
	}
	*hs = HaikuState(str)
	return nil
}

// Value for HaikuState
func (hs HaikuState) Value() (driver.Value, error) {
	return string(hs), nil
}
