package sensor

import (
	"github.com/GeoNet/kit/gloria_pb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var s3Client *s3.S3

func init() {
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		log.Print(err)
	}
	s3Client = s3.New(s)
}

func gloria(in io.Reader) (locations, error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return locations{}, err
	}

	var m gloria_pb.Marks

	err = proto.Unmarshal(b, &m)
	if err != nil {
		return locations{}, err
	}

	var l locations

	for _, v := range m.Marks {

		f := locationFeature{
			Type: "Feature",
			Properties: location{
				Mark: v.GetCode(),
				Code: v.GetCode(),
			},
			Geometry: point{
				Type: "Point",
				Coordinates: [2]float64{
					v.GetPoint().GetLongitude(),
					v.GetPoint().GetLatitude(),
				},
			},
		}

		for _, a := range v.GetInstalledAntenna() {
			f.Properties.Channels = append(f.Properties.Channels, channel{
				SensorType: "GNSS Antenna",
				Start:      time.Unix(a.GetSpan().GetStart(), 0),
				End:        time.Unix(a.GetSpan().GetEnd(), 0),
			})
		}

		l = append(l, f)
	}

	return l, nil
}

//sensorStart before searchEnd AND sensorEnd after searchStart
func gloriaSensorGeoJSON(sensorCode string, stationCode string, start time.Time, end time.Time, g *LocationFeatures) error {
	r, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("geonet-meta"),
		Key:    aws.String("config/gloria/mark-index.pb"),
	})
	if err != nil {
		return err
	}
	defer r.Body.Close()

	c, err := gloria(r.Body)
	if err != nil {
		return err
	}

	locations := c.sensor(sensorCode).channel(start, end, sensorCode)
	for _, f := range locations {
		f.Properties.Channels = []channel{}
		f.Properties.SensorType = sensorCode

		match := strings.HasPrefix(strings.ToLower(f.Properties.Code), strings.ToLower(stationCode)) // same as `regexp.MatchString(strings.ToLower(stationCode)+".*", strings.ToLower(f.Properties.Code))`?
		if stationCode == "" || match {
			g.Features = append(g.Features, f)
		}
	}
	return err
}
