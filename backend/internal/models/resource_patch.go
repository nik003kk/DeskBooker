package models

// ResourcePatch can update only editable resource fields. A nil pointer
// means omitted; FloorSet and Features distinguish omitted from explicit null/[].
type ResourcePatch struct {
    Name *string
    Type *ResourceType
    Floor *int
    FloorSet bool
    Capacity *int
    IsActive *bool
    Features []string
    FeaturesSet bool
}
