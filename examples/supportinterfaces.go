package examples

import "github.com/av-belyakov/objectsmispformat"

type SpecialObjectComparator interface {
	ComparisonID(string) bool
	ComparisonEvent(*objectsmispformat.EventsMispFormat) bool
	ComparisonReports(*objectsmispformat.EventReports) bool
	ComparisonAttributes([]*objectsmispformat.AttributesMispFormat) bool
	ComparisonObjects(map[int]*objectsmispformat.ObjectsMispFormat) bool
	ComparisonObjectTags(*objectsmispformat.ListEventObjectTags) bool
	SpecialObjectGetter
}

type SpecialObjectGetter interface {
	GetID() string
	GetEvent() *objectsmispformat.EventsMispFormat
	GetReports() *objectsmispformat.EventReports
	GetAttributes() []*objectsmispformat.AttributesMispFormat
	GetObjects() map[int]*objectsmispformat.ObjectsMispFormat
	GetObjectTags() *objectsmispformat.ListEventObjectTags
}
