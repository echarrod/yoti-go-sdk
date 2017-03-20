package yoti

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/getyoti/go/attrpubapi_v1"
	"github.com/getyoti/go/compubapi_v1"
	"github.com/golang/protobuf/proto"
)

const apiUrl = "https://api.yoti.com/api/v1"

// YotiClient represents a client that can communicate with yoti and return information about Yoti users.
type YotiClient struct {
	// SdkID represents the SDK ID and NOT the App ID. This can be found in the integration section of your
	// application dashboard at https://www.yoti.com/dashboard/
	SdkID string

	// Key should be the security key given to you by yoti (see: security keys section of
	// https://www.yoti.com/dashboard/) for more information about how to load your key from a file see:
	// https://github.com/getyoti/go/blob/master/README.md
	Key []byte
}

// GetActivityDetails requests information about a Yoti user using the token generated by the Yoti login process.
// It returns the outcome of the request. If the request was successful it will include the users details, otherwise
// it will specify a reason the request failed.
func (client *YotiClient) GetUserProfile(token string) (YotiUserProfile, error) {
	return getActivityDetails(doRequest, token, client.SdkID, client.Key)
}

func getActivityDetails(requester httpRequester, encryptedToken, sdkId string, keyBytes []byte) (result YotiUserProfile, err error) {
	var key *rsa.PrivateKey
	if key, err = loadRsaKey(keyBytes); err != nil {
		err = fmt.Errorf("Invalid Key: %s", err.Error())
		return
	}

	// query parameters
	var token string
	if token, err = decryptToken(encryptedToken, key); err != nil {
		return
	}

	var nonce string
	if nonce, err = generateNonce(); err != nil {
		return
	}

	timestamp := getTimestamp()

	// create http endpoint
	endpoint := getEndpoint(token, nonce, timestamp, sdkId)

	// create request headers
	var authKey string
	if authKey, err = getAuthKey(key); err != nil {
		return
	}

	var authDigest string
	if authDigest, err = getAuthDigest(endpoint, key); err != nil {
		return
	}

	headers := make(map[string]string)

	headers["X-Yoti-Auth-Key"] = authKey
	headers["X-Yoti-Auth-Digest"] = authDigest

	var response *httpResponse
	if response, err = requester(apiUrl+endpoint, headers); err != nil {
		return
	}

	if response.Success {
		var parsedResponse = profileDO{}

		if err = json.Unmarshal([]byte(response.Content), &parsedResponse); err != nil {
			return
		}

		if parsedResponse.Receipt.SharingOutcome != "SUCCESS" {
			err = errors.New(ActivityOutcome_SharingFailure)
		} else {
			var attributeList *attrpubapi_v1.AttributeList
			if attributeList, err = decryptCurrentUserReceipt(&parsedResponse.Receipt, key); err != nil {
				return
			}

			id := parsedResponse.Receipt.RememberMeId

			result = YotiUserProfile{
				ID:              id,
				OtherAttributes: make(map[string]YotiAttributeValue)}

			if attributeList == nil {
				return
			}

			for _, attribute := range attributeList.Attributes {
				switch attribute.Name {
				case "selfie":
					data := make([]byte, len(attribute.Value))
					copy(data, attribute.Value)

					switch attribute.ContentType {
					case attrpubapi_v1.ContentType_JPEG:
						result.Selfie = &Image{
							Type: ImageType_Jpeg,
							Data: data}
					case attrpubapi_v1.ContentType_PNG:
						result.Selfie = &Image{
							Type: ImageType_Png,
							Data: data}
					}
				case "given_names":
					result.GivenNames = string(attribute.Value)
				case "family_name":
					result.FamilyName = string(attribute.Value)
				case "phone_number":
					result.MobileNumber = string(attribute.Value)
				case "date_of_birth":
					parsedTime, err := time.Parse("2006-01-02", string(attribute.Value))
					if err == nil {
						result.DateOfBirth = &parsedTime
					}
				case "post_code":
					result.Address = string(attribute.Value)
				case "gender":
					result.Gender = string(attribute.Value)
				case "nationality":
					result.Nationality = string(attribute.Value)
				default:
					switch attribute.ContentType {
					case attrpubapi_v1.ContentType_DATE:
						result.OtherAttributes[attribute.Name] = YotiAttributeValue{
							Type:  AttributeType_Date,
							Value: attribute.Value}
					case attrpubapi_v1.ContentType_STRING:
						result.OtherAttributes[attribute.Name] = YotiAttributeValue{
							Type:  AttributeType_Text,
							Value: attribute.Value}
					case attrpubapi_v1.ContentType_JPEG:
						result.OtherAttributes[attribute.Name] = YotiAttributeValue{
							Type:  AttributeType_Jpeg,
							Value: attribute.Value}
					case attrpubapi_v1.ContentType_PNG:
						result.OtherAttributes[attribute.Name] = YotiAttributeValue{
							Type:  AttributeType_Png,
							Value: attribute.Value}
					}
				}
			}
		}
	} else {
		switch response.StatusCode {
		case http.StatusNotFound:
			err = errors.New(ActivityOutcome_ProfileNotFound)
		default:
			err = errors.New(ActivityOutcome_Failure)
		}
	}

	return
}

func decryptCurrentUserReceipt(receipt *receiptDO, key *rsa.PrivateKey) (result *attrpubapi_v1.AttributeList, err error) {
	var unwrappedKey []byte
	if unwrappedKey, err = unwrapKey(receipt.WrappedReceiptKey, key); err != nil {
		return
	}

	if receipt.OtherPartyProfileContent == "" {
		return
	}

	var otherPartyProfileContentBytes []byte
	if otherPartyProfileContentBytes, err = base64ToBytes(receipt.OtherPartyProfileContent); err != nil {
		return
	}

	encryptedData := &compubapi_v1.EncryptedData{}
	if err = proto.Unmarshal(otherPartyProfileContentBytes, encryptedData); err != nil {
		return nil, err
	}

	var decipheredBytes []byte
	if decipheredBytes, err = decipherAes(unwrappedKey, encryptedData.Iv, encryptedData.CipherText); err != nil {
		return nil, err
	}

	attributeList := &attrpubapi_v1.AttributeList{}
	if err := proto.Unmarshal(decipheredBytes, attributeList); err != nil {
		return nil, err
	}

	return attributeList, nil
}

func getAuthKey(key *rsa.PrivateKey) (string, error) {
	return getDerEncodedPublicKey(key)
}

func getEndpoint(token, nonce, timestamp, sdkId string) string {
	return fmt.Sprintf("/profile/%s?nonce=%s&timestamp=%s&appId=%s", token, nonce, timestamp, sdkId)
}

func getAuthDigest(endpoint string, key *rsa.PrivateKey) (result string, err error) {
	digestBytes := utfToBytes("GET&" + endpoint)

	var signedDigestBytes []byte

	if signedDigestBytes, err = signDigest(digestBytes, key); err != nil {
		return
	}

	result = bytesToBase64(signedDigestBytes)
	return
}

func getTimestamp() string {
	return strconv.FormatInt(time.Now().Unix()*1000, 10)
}

func decryptToken(encryptedConnectToken string, key *rsa.PrivateKey) (result string, err error) {
	// token was encoded as a urlsafe base64 so it can be transfered in a url
	var cipherBytes []byte
	if cipherBytes, err = urlSafeBase64ToBytes(encryptedConnectToken); err != nil {
		return
	}

	var decipheredBytes []byte
	if decipheredBytes, err = decryptRsa(cipherBytes, key); err != nil {
		return
	}

	result = bytesToUtf8(decipheredBytes)
	return
}

func unwrapKey(wrappedKey string, key *rsa.PrivateKey) (result []byte, err error) {
	var cipherBytes []byte
	if cipherBytes, err = base64ToBytes(wrappedKey); err != nil {
		return
	}
	result, err = decryptRsa(cipherBytes, key)
	return
}
