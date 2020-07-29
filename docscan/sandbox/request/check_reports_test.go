package request

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/getyoti/yoti-go-sdk/v3/docscan/sandbox/request/check"
	"github.com/getyoti/yoti-go-sdk/v3/docscan/sandbox/request/check/report"
)

func ExampleCheckReportsBuilder() {
	breakdown, err := report.NewBreakdownBuilder().
		WithResult("some_result").
		WithSubCheck("some_check").
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	recommendation, err := report.NewRecommendationBuilder().
		WithValue("some_value").
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	authenticityCheck, err := check.NewDocumentAuthenticityCheckBuilder().
		WithBreakdown(breakdown).
		WithRecommendation(recommendation).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	faceMatchCheck, err := check.NewDocumentFaceMatchCheckBuilder().
		WithBreakdown(breakdown).
		WithRecommendation(recommendation).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	textDataCheck, err := check.NewDocumentTextDataCheckBuilder().
		WithBreakdown(breakdown).
		WithRecommendation(recommendation).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	zoomLivenessCheck, err := check.NewZoomLivenessCheckBuilder().Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	checkReports, err := NewCheckReportsBuilder().
		WithDocumentAuthenticityCheck(authenticityCheck).
		WithDocumentFaceMatchCheck(faceMatchCheck).
		WithDocumentTextDataCheck(textDataCheck).
		WithLivenessCheck(zoomLivenessCheck).
		WithAsyncReportDelay(10).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	data, _ := json.Marshal(checkReports)
	fmt.Println(string(data))
	// Output: {"ID_DOCUMENT_AUTHENTICITY":[{"result":{"report":{"recommendation":{"value":"some_value"},"breakdown":[{"sub_check":"some_check","result":"some_result","details":[]}]}}}],"ID_DOCUMENT_TEXT_DATA_CHECK":[{"result":{"report":{"recommendation":{"value":"some_value"},"breakdown":[{"sub_check":"some_check","result":"some_result","details":[]}]}}}],"ID_DOCUMENT_FACE_MATCH":[{"result":{"report":{"recommendation":{"value":"some_value"},"breakdown":[{"sub_check":"some_check","result":"some_result","details":[]}]}}}],"LIVENESS":[{"result":{"report":{}},"liveness_type":"ZOOM"}],"async_report_delay":10}
}

func ExampleCheckReportsBuilder_minimal() {
	checkReports, err := NewCheckReportsBuilder().Build()
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: %s", err.Error())
		return
	}

	data, _ := json.Marshal(checkReports)
	fmt.Println(string(data))
	// Output: {"ID_DOCUMENT_AUTHENTICITY":[],"ID_DOCUMENT_TEXT_DATA_CHECK":[],"ID_DOCUMENT_FACE_MATCH":[],"LIVENESS":[]}
}
