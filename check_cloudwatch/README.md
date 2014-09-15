# check\_cloudwatch

## Description

Check for Amazon CloudWatch Metrics

## Usage

    Usage: check_cloudwatch [options]
      -critical=0: Critical threshold
      -dimension=[]: The dimensions of the metric
      -metric-name="": The name of the metric
      -namespace="": The namespace of the metric
      -period=60: The length in seconds for aggregation
      -region="": AWS region
      -statistic="": The statistic of the metric
      -warning=0: Warning threshold

## Example

    check_cloudwatch \
        --namespace AWS/EC2 \
        --metric-name CPUUtilization \
        --statistic Average \
        --period 360 \
        --region=ap-northeast-1 \
        --dimension InstanceId=i-xxxxxxxx \
        --warning 90 \
        --critical 95

## Install

    go get github.com/kaorimatz/nagios-plugins/check_cloudwatch
