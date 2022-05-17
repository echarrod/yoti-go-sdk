package retrieve_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/getyoti/yoti-go-sdk/v3/docscan/constants"
	"github.com/getyoti/yoti-go-sdk/v3/docscan/session/retrieve"
	"github.com/getyoti/yoti-go-sdk/v3/file"
	"gotest.tools/v3/assert"
)

func TestGetSessionResult_UnmarshalJSON(t *testing.T) {
	authenticityCheckResponse := &retrieve.CheckResponse{
		Type:  constants.IDDocumentAuthenticity,
		State: "DONE",
	}

	testDate := time.Date(2020, 01, 01, 1, 2, 3, 4, time.UTC)
	faceMatchCheckResponse := &retrieve.CheckResponse{
		Type:    constants.IDDocumentFaceMatch,
		Created: &testDate,
	}

	textDataCheckResponse := &retrieve.CheckResponse{
		Type:   constants.IDDocumentTextDataCheck,
		Report: &retrieve.ReportResponse{},
	}

	livenessCheckResponse := &retrieve.CheckResponse{
		Type:        constants.Liveness,
		LastUpdated: &testDate,
	}

	idDocComparisonCheckResponse := &retrieve.CheckResponse{
		Type:  constants.IDDocumentComparison,
		State: "PENDING",
	}

	thirdPartyIdentityCheckResponse := &retrieve.CheckResponse{
		Type:  constants.ThirdPartyIdentityCheck,
		State: "PENDING",
	}

	supplementaryTextDataCheckResponse := &retrieve.CheckResponse{
		Type:   constants.SupplementaryDocumentTextDataCheck,
		Report: &retrieve.ReportResponse{},
	}

	watchlistScreeningCheckResponse := &retrieve.CheckResponse{
		Type:  constants.WatchlistScreening,
		State: "DONE",
	}

	var checks []*retrieve.CheckResponse
	checks = append(checks, &retrieve.CheckResponse{Type: "OTHER_TYPE", ID: "id"})
	checks = append(checks, authenticityCheckResponse)
	checks = append(checks, faceMatchCheckResponse)
	checks = append(checks, textDataCheckResponse)
	checks = append(checks, livenessCheckResponse)
	checks = append(checks, idDocComparisonCheckResponse)
	checks = append(checks, thirdPartyIdentityCheckResponse)
	checks = append(checks, supplementaryTextDataCheckResponse)
	checks = append(checks, watchlistScreeningCheckResponse)

	biometricConsentTimestamp := time.Date(2020, 01, 01, 1, 2, 3, 4, time.UTC)

	getSessionResult := retrieve.GetSessionResult{
		Checks:                    checks,
		BiometricConsentTimestamp: &biometricConsentTimestamp,
	}
	marshalled, err := json.Marshal(&getSessionResult)
	assert.NilError(t, err)

	var result retrieve.GetSessionResult
	err = json.Unmarshal(marshalled, &result)
	assert.NilError(t, err)

	assert.Equal(t, 9, len(result.Checks))

	assert.Equal(t, 1, len(result.AuthenticityChecks()))
	assert.Equal(t, "DONE", result.AuthenticityChecks()[0].State)

	assert.Equal(t, 1, len(result.FaceMatchChecks()))
	assert.Check(t, result.FaceMatchChecks()[0].Created.Equal(testDate))

	assert.Equal(t, 1, len(result.TextDataChecks()))
	assert.DeepEqual(t, &retrieve.ReportResponse{}, result.TextDataChecks()[0].Report)

	assert.Equal(t, 1, len(result.IDDocumentTextDataChecks()))
	assert.DeepEqual(t, &retrieve.ReportResponse{}, result.IDDocumentTextDataChecks()[0].Report)

	assert.Equal(t, 1, len(result.LivenessChecks()))
	assert.Check(t, result.LivenessChecks()[0].LastUpdated.Equal(testDate))

	assert.Equal(t, 1, len(result.IDDocumentComparisonChecks()))
	assert.Equal(t, "PENDING", result.IDDocumentComparisonChecks()[0].State)

	assert.Equal(t, 1, len(result.ThirdPartyIdentityChecks()))
	assert.Equal(t, "PENDING", result.ThirdPartyIdentityChecks()[0].State)

	assert.Equal(t, 1, len(result.SupplementaryDocumentTextDataChecks()))
	assert.DeepEqual(t, &retrieve.ReportResponse{}, result.SupplementaryDocumentTextDataChecks()[0].Report)

	assert.Equal(t, 1, len(result.WatchlistScreeningChecks()))
	assert.DeepEqual(t, "DONE", result.WatchlistScreeningChecks()[0].State)

	assert.Equal(t, biometricConsentTimestamp, *result.BiometricConsentTimestamp)
}

func TestGetSessionResult_UnmarshalJSON_Watchlist(t *testing.T) {
	bytes, err := file.ReadFile("../../../test/fixtures/watchlist_screening.json")
	assert.NilError(t, err)

	var result retrieve.GetSessionResult
	err = result.UnmarshalJSON(bytes)
	assert.NilError(t, err)

	assert.Equal(t, 1, len(result.WatchlistScreeningChecks()))
	watchlistScreeningCheck := result.WatchlistScreeningChecks()[0]
	assert.Equal(t, watchlistScreeningCheck.GeneratedProfile.Media.Type, "JSON")

	watchlistSummary := watchlistScreeningCheck.Report.WatchlistSummary

	assert.Equal(t, 0, watchlistSummary.TotalHits)
	assert.Equal(t, 2, len(watchlistSummary.SearchConfig.Categories))
	assert.Equal(t, watchlistSummary.SearchConfig.Categories[0], "ADVERSE-MEDIA")
	assert.Equal(t, watchlistSummary.SearchConfig.Categories[1], "SANCTIONS")
	assert.Equal(t, watchlistSummary.RawResults.Media.Type, "JSON")
	assert.Equal(t, watchlistSummary.AssociatedCountryCodes[0], "GBR")
}

func TestGetSessionResult_UnmarshalJSON_Invalid(t *testing.T) {
	var result retrieve.GetSessionResult
	err := result.UnmarshalJSON([]byte("some-invalid-json"))
	assert.ErrorContains(t, err, "invalid character")
}

func TestGetSessionResult_UnmarshalJSON_WithoutBiometricConsentTimestamp(t *testing.T) {
	var result retrieve.GetSessionResult
	err := result.UnmarshalJSON([]byte("{}"))
	assert.NilError(t, err)
	assert.Check(t, result.BiometricConsentTimestamp == nil)
}
