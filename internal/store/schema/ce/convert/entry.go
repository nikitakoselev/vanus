// Copyright 2022 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convert

import (
	// standard libraries.
	"sort"
	"strconv"
	"time"

	// third-party libraries.
	cepb "cloudevents.io/genproto/v1"
	"google.golang.org/protobuf/proto"

	// first-party libraries.
	segpb "github.com/linkall-labs/vanus/proto/pkg/segment"

	// this project.
	"github.com/linkall-labs/vanus/internal/store/block"
	ceschema "github.com/linkall-labs/vanus/internal/store/schema/ce"
)

const (
	dataContentTypeAttr = "datacontenttype"
	dataSchemaAttr      = "dataschema"
	subjectAttr         = "subject"
	timeAttr            = "time"
)

type ceEntry struct {
	block.EmptyEntry
	ce *cepb.CloudEvent
}

// Make sure ceEntry implements block.EntryExt.
var _ block.EntryExt = (*ceEntry)(nil)

func (e *ceEntry) GetBytes(ordinal int) []byte {
	if ordinal != ceschema.DataOrdinal {
		return nil
	}

	switch data := e.ce.Data.(type) {
	case *cepb.CloudEvent_BinaryData:
		return data.BinaryData
	case *cepb.CloudEvent_TextData:
		return []byte(data.TextData)
	case *cepb.CloudEvent_ProtoData:
		buf, _ := proto.Marshal(data.ProtoData)
		return buf
	}
	return nil
}

func (e *ceEntry) GetString(ordinal int) string {
	switch ordinal {
	case ceschema.IDOrdinal:
		return e.ce.Id
	case ceschema.SourceOrdinal:
		return e.ce.Source
	case ceschema.SpecVersionOrdinal:
		return e.ce.SpecVersion
	case ceschema.TypeOrdinal:
		return e.ce.Type
	}

	if e.ce.Attributes == nil {
		return ""
	}

	var attr *cepb.CloudEventAttributeValue
	switch ordinal {
	case ceschema.DataContentTypeOrdinal:
		attr = e.ce.Attributes[dataContentTypeAttr]
	case ceschema.DataSchemaOrdinal:
		attr = e.ce.Attributes[dataSchemaAttr]
	case ceschema.SubjectOrdinal:
		attr = e.ce.Attributes[subjectAttr]
	}
	return attr.GetCeString()
}

func (e *ceEntry) GetTime(ordinal int) time.Time {
	if ordinal != ceschema.TimeOrdinal || e.ce.Attributes == nil {
		return time.Time{}
	}
	attr := e.ce.Attributes[timeAttr]
	return attr.GetCeTimestamp().AsTime()
}

func (e *ceEntry) RangeOptionalAttributes(cb block.OptionalAttributeCallback) {
	// id, source, specversion, type, datacontenttype, dataschema, subject, time
	cb.OnString(ceschema.IDOrdinal, e.ce.Id)
	cb.OnString(ceschema.SourceOrdinal, e.ce.Source)
	cb.OnString(ceschema.SpecVersionOrdinal, e.ce.SpecVersion)
	cb.OnString(ceschema.TypeOrdinal, e.ce.Type)
	if e.ce.Data != nil {
		switch data := e.ce.Data.(type) {
		case *cepb.CloudEvent_BinaryData:
			cb.OnBytes(ceschema.DataOrdinal, data.BinaryData)
		case *cepb.CloudEvent_TextData:
			cb.OnString(ceschema.DataOrdinal, data.TextData)
		case *cepb.CloudEvent_ProtoData:
			// TODO(james.yin): TypeUrl
			cb.OnBytes(ceschema.DataOrdinal, data.ProtoData.Value)
		}
	}
	if e.ce.Attributes != nil {
		if v, ok := e.ce.Attributes[dataContentTypeAttr]; ok {
			cb.OnString(ceschema.DataContentTypeOrdinal, v.GetCeString())
		}
		if v, ok := e.ce.Attributes[dataSchemaAttr]; ok {
			cb.OnString(ceschema.DataSchemaOrdinal, v.GetCeString())
		}
		if v, ok := e.ce.Attributes[subjectAttr]; ok {
			cb.OnString(ceschema.SubjectOrdinal, v.GetCeString())
		}
		if v, ok := e.ce.Attributes[timeAttr]; ok {
			cb.OnTime(ceschema.TimeOrdinal, v.GetCeTimestamp().AsTime())
		}
	}
}

func (e *ceEntry) OptionalAttributeCount() int {
	sz := 4
	if e.ce.Data != nil {
		sz++
	}
	if e.ce.Attributes != nil {
		for _, attr := range []string{dataContentTypeAttr, dataSchemaAttr, subjectAttr, timeAttr} {
			if _, ok := e.ce.Attributes[attr]; ok {
				sz++
			}
		}
	}
	return sz
}

func (e *ceEntry) GetExtensionAttribute(attr []byte) []byte {
	if e.ce.Attributes == nil {
		return nil
	}
	if v, ok := e.ce.Attributes[string(attr)]; ok {
		return attrValue(v)
	}
	return nil
}

func (e *ceEntry) RangeExtensionAttributes(cb block.ExtensionAttributeCallback) {
	if len(e.ce.Attributes) == 0 {
		return
	}

	// Make sure the order of attributes.
	attrs := make([]string, 0, len(e.ce.Attributes))
	for attr := range e.ce.Attributes {
		switch attr {
		case dataContentTypeAttr, dataSchemaAttr, subjectAttr, timeAttr:
		default:
			attrs = append(attrs, attr)
		}
	}
	sort.Strings(attrs)

	for _, attr := range attrs {
		cb.OnAttribute([]byte(attr), attrValue(e.ce.Attributes[attr]))
	}
}

func (e *ceEntry) ExtensionAttributeCount() int {
	sz := len(e.ce.Attributes)
	if sz == 0 {
		return 0
	}
	for _, attr := range []string{dataContentTypeAttr, dataSchemaAttr, subjectAttr, timeAttr} {
		if _, ok := e.ce.Attributes[attr]; ok {
			sz--
		}
	}
	return sz
}

func attrValue(v *cepb.CloudEventAttributeValue) []byte {
	// FIXME(james.yin): support native types.
	switch val := v.GetAttr().(type) {
	case *cepb.CloudEventAttributeValue_CeBoolean:
		return []byte(strconv.FormatBool(val.CeBoolean))
	case *cepb.CloudEventAttributeValue_CeInteger:
		return []byte(strconv.FormatInt(int64(val.CeInteger), 10))
	case *cepb.CloudEventAttributeValue_CeString:
		return []byte(val.CeString)
	case *cepb.CloudEventAttributeValue_CeBytes:
		return val.CeBytes
	case *cepb.CloudEventAttributeValue_CeUri:
		return []byte(val.CeUri)
	case *cepb.CloudEventAttributeValue_CeUriRef:
		return []byte(val.CeUriRef)
	case *cepb.CloudEventAttributeValue_CeTimestamp:
		return []byte(val.CeTimestamp.AsTime().Format(time.RFC3339Nano))
	}
	return nil
}

func ToEntry(event *cepb.CloudEvent) block.EntryExt {
	if len(event.Attributes) != 0 {
		delete(event.Attributes, segpb.XVanusBlockOffset)
		delete(event.Attributes, segpb.XVanusLogOffset)
		delete(event.Attributes, segpb.XVanusStime)
	}
	return &ceEntry{ce: event}
}
