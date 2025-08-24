package influx

import (
	"fmt"
	"sort"
	"strings"
)

// Data represents data to be sent to InfluxDB
type Data struct {
	Timestamp int64
	Name      string
	Bucket    string
	Tags      map[string]string
	Fields    map[string]string
}

// New creates a new InfluxData struct
func New() *Data {
	return &Data{
		Tags:   make(map[string]string),
		Fields: make(map[string]string),
	}
}

// Marshal converts InfluxData into Influx wire protocol
func (m *Data) Marshal() string {
	tags := make([]string, 0, len(m.Tags))
	for tag, value := range m.Tags {
		tags = append(tags, tag+"="+value)
	}
	sort.Strings(tags)

	fields := make([]string, 0, len(m.Fields))
	for field, value := range m.Fields {
		fields = append(fields, field+"="+value)
	}
	sort.Strings(fields)

	return fmt.Sprintf("%s,%s %s %d\n",
		m.Name,
		strings.Join(tags, ","),
		strings.Join(fields, ","),
		m.Timestamp)
}
