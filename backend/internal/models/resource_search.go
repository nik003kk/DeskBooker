package models
import "time"

// ResourceSearch carries validated user search filters.
type ResourceSearch struct {
 Type *ResourceType
 StartsAt *time.Time
 EndsAt *time.Time
 Features []string
}
