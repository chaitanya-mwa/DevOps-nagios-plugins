package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/cloudwatch"
)

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type options struct {
	criticalThreshold float64
	dimensions        []cloudwatch.Dimension
	metricName        string
	namespace         string
	period            int
	region            aws.Region
	statistic         string
	warningThreshold  float64
}

func ok(metricName, message string) {
	fmt.Printf("%s OK - %s\n", metricName, message)
	os.Exit(OK)
}

func warning(metricName, message string) {
	fmt.Printf("%s WARNING - %s\n", metricName, message)
	os.Exit(WARNING)
}

func critical(metricName, message string) {
	fmt.Printf("%s CRITICAL - %s\n", metricName, message)
	os.Exit(CRITICAL)
}

func unknown(metricName, message string) {
	fmt.Printf("%s UNKNOWN - %s\n", metricName, message)
	os.Exit(UNKNOWN)
}

func auth() (aws.Auth, error) {
	return aws.GetAuth("", "", "", time.Time{})
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: check_cloudwatch [options]\n")
	flag.PrintDefaults()
}

func parseCommandLine() *options {
	var options options
	flag.Usage = usage
	flag.Float64Var(&options.criticalThreshold, "critical", 0, "Critical threshold")
	flag.Var(newDimensionsValue([]cloudwatch.Dimension{}, &options.dimensions), "dimension", "The dimensions of the metric")
	flag.StringVar(&options.metricName, "metric-name", "", "The name of the metric")
	flag.StringVar(&options.namespace, "namespace", "", "The namespace of the metric")
	flag.IntVar(&options.period, "period", 60, "The length in seconds for aggregation")
	flag.Var(newRegionValue(aws.Region{}, &options.region), "region", "AWS region")
	flag.StringVar(&options.statistic, "statistic", "", "The statistic of the metric")
	flag.Float64Var(&options.warningThreshold, "warning", 0, "Warning threshold")
	flag.Parse()
	return &options
}

type dimensionsValue []cloudwatch.Dimension

func newDimensionsValue(value []cloudwatch.Dimension, p *[]cloudwatch.Dimension) *dimensionsValue {
	*p = value
	return (*dimensionsValue)(p)
}

func (d *dimensionsValue) Set(value string) error {
	nameValue := strings.SplitN(value, "=", 2)
	dimension := cloudwatch.Dimension{
		Name:  nameValue[0],
		Value: nameValue[1],
	}
	*d = append(*d, dimension)
	return nil
}

func (d *dimensionsValue) String() string {
	nameValues := make([]string, len(*d))
	for i, dimension := range *d {
		nameValues[i] = fmt.Sprintf("%s=%s", dimension.Name, dimension.Value)
	}
	return fmt.Sprintf("%v", nameValues)
}

type regionValue aws.Region

func newRegionValue(value aws.Region, p *aws.Region) *regionValue {
	*p = value
	return (*regionValue)(p)
}

func (r *regionValue) Set(value string) error {
	*r = regionValue(aws.Regions[value])
	return nil
}

func (r *regionValue) String() string {
	return fmt.Sprintf("%q", (*aws.Region)(r).Name)
}

func getData(datapoint cloudwatch.Datapoint, statistic string) (float64, error) {
	switch statistic {
	case "Minimum":
		return datapoint.Minimum, nil
	case "Maximum":
		return datapoint.Maximum, nil
	case "Sum":
		return datapoint.Sum, nil
	case "Average":
		return datapoint.Average, nil
	case "SampleCount":
		return datapoint.SampleCount, nil
	default:
		return 0, fmt.Errorf("Unknown statistic: %s", statistic)
	}
}

func main() {
	options := parseCommandLine()

	auth, err := auth()
	if err != nil {
		unknown(options.metricName, err.Error())
	}

	client, err := cloudwatch.NewCloudWatch(auth, options.region.CloudWatchServicepoint)
	if err != nil {
		unknown(options.metricName, err.Error())
	}

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Duration(options.period) * time.Second)
	request := &cloudwatch.GetMetricStatisticsRequest{
		Namespace:  options.namespace,
		MetricName: options.metricName,
		Dimensions: options.dimensions,
		StartTime:  startTime,
		EndTime:    endTime,
		Period:     options.period,
		Statistics: []string{options.statistic},
	}
	response, err := client.GetMetricStatistics(request)
	if err != nil {
		unknown(options.metricName, err.Error())
	}

	datapoints := response.GetMetricStatisticsResult.Datapoints
	if len(datapoints) == 0 {
		unknown(options.metricName, "No datapoints")
	}
	datapoint := datapoints[0]
	data, err := getData(datapoint, options.statistic)
	if err != nil {
		unknown(options.metricName, err.Error())
	}
	unit := datapoint.Unit

	message := fmt.Sprintf("%f %s", data, unit)
	if options.criticalThreshold > options.warningThreshold {
		if data > options.criticalThreshold {
			critical(options.metricName, message)
		} else if data > options.warningThreshold {
			warning(options.metricName, message)
		}
	} else {
		if data < options.criticalThreshold {
			critical(options.metricName, message)
		} else if data < options.warningThreshold {
			warning(options.metricName, message)
		}
	}
	ok(options.metricName, message)
}
