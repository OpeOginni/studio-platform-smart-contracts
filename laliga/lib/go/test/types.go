package test

import (
	"fmt"
	"github.com/onflow/cadence"
)

type SeriesData struct {
	ID     uint64
	Name   string
	Active bool
}
type SetData struct {
	ID     uint64
	Name   string
	Locked bool
}
type PlayData struct {
	ID             uint64
	Classification string
	Metadata       map[string]string
}
type EditionData struct {
	ID          uint64
	SeriesID    uint64
	SetID       uint64
	PlayID      uint64
	MaxMintSize *uint64
	Tier        string
}
type OurNFTData struct {
	ID           uint64
	EditionID    uint64
	SerialNumber uint64
	// A UFix64 in uint64 form
	MintingDate uint64
}

type DisplayView struct {
	Name        string
	Description string
	ImageURL    string
}

type EditionView struct {
	Number uint64
	Max    uint64
}

type NFTCollectionDataView struct {
	StoragePath        string
	PublicPath         string
	ProviderPath       string
	PublicCollection   string
	PublicLinkedType   string
	ProviderLinkedType string
}

type TraitView struct {
	Name  string
	Value string
}

type TraitsView []TraitView

func cadenceStringDictToGo(cadenceDict cadence.Dictionary) map[string]string {
	goDict := make(map[string]string)
	for _, pair := range cadenceDict.Pairs {
		goDict[pair.Key.ToGoValue().(string)] = pair.Value.ToGoValue().(string)
	}
	return goDict
}

func parseSeriesData(value cadence.Value) SeriesData {
	fields := value.(cadence.Struct).Fields
	return SeriesData{
		fields[0].ToGoValue().(uint64),
		fields[1].ToGoValue().(string),
		fields[2].ToGoValue().(bool),
	}
}

func parseSetData(value cadence.Value) SetData {
	fields := value.(cadence.Struct).Fields
	return SetData{
		fields[0].ToGoValue().(uint64),
		fields[1].ToGoValue().(string),
		fields[2].ToGoValue().(bool),
	}
}

func parsePlayData(value cadence.Value) PlayData {
	fields := value.(cadence.Struct).Fields
	return PlayData{
		fields[0].ToGoValue().(uint64),
		fields[1].ToGoValue().(string),
		cadenceStringDictToGo(fields[2].(cadence.Dictionary)),
	}
}

func parseEditionData(value cadence.Value) EditionData {
	fields := value.(cadence.Struct).Fields
	var maxMintSize uint64
	if fields[4] != nil && fields[4].ToGoValue() != nil {
		maxMintSize = fields[4].ToGoValue().(uint64)
	}
	return EditionData{
		fields[0].ToGoValue().(uint64),
		fields[1].ToGoValue().(uint64),
		fields[2].ToGoValue().(uint64),
		fields[3].ToGoValue().(uint64),
		&maxMintSize,
		fields[5].ToGoValue().(string),
	}
}

func parseNFTProperties(value cadence.Value) OurNFTData {
	array := value.(cadence.Array).Values
	return OurNFTData{
		array[0].ToGoValue().(uint64),
		array[1].ToGoValue().(uint64),
		array[2].ToGoValue().(uint64),
		array[3].ToGoValue().(uint64),
	}
}

func parseMetadataDisplayView(value cadence.Value) DisplayView {
	fields := value.(cadence.Struct).Fields
	return DisplayView{
		fields[0].ToGoValue().(string),
		fields[1].ToGoValue().(string),
		fields[2].ToGoValue().(string),
	}
}

func parseMetadataEditionView(value cadence.Value) EditionView {
	fields := value.(cadence.Struct).Fields
	edition := fields[0].(cadence.Array).Values[0].(cadence.Struct).Fields
	maxMintSize := uint64(0)
	if edition[2].ToGoValue() != nil {
		maxMintSize = edition[2].ToGoValue().(uint64)
	}
	return EditionView{
		edition[1].ToGoValue().(uint64),
		maxMintSize,
	}
}

func parseMetadataSerialView(value cadence.Value) uint64 {
	return value.ToGoValue().(uint64)
}

func parseMetadataNFTCollectionDataView(value cadence.Value) NFTCollectionDataView {
	fields := value.(cadence.Struct).Fields
	return NFTCollectionDataView{
		StoragePath:        fields[0].(cadence.Path).Identifier,
		PublicPath:         fields[1].(cadence.Path).Identifier,
		ProviderPath:       fields[2].(cadence.Path).Identifier,
		PublicCollection:   fields[3].(cadence.TypeValue).StaticType.ID(),
		PublicLinkedType:   fields[4].(cadence.TypeValue).StaticType.ID(),
		ProviderLinkedType: fields[5].(cadence.TypeValue).StaticType.ID(),
	}
}

func parseMetadataTraitsView(value cadence.Value) TraitsView {
	var view TraitsView
	fields := value.(cadence.Struct).Fields
	for _, val := range fields[0].(cadence.Array).Values {
		trait := val.(cadence.Struct).Fields
		view = append(view, TraitView{
			Name:  fmt.Sprintf("%v", trait[0].ToGoValue()),
			Value: fmt.Sprintf("%v", trait[1].ToGoValue()),
		})
	}
	return view
}
