package tx

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"
)

const apiURL = "tmt.tencentcloudapi.com"
const apiID = ""
const apiKey = ""

const qps = 5
const interval = int(1000.0 / float32(qps))
const LengthLimit = 2000

var timeCostMilliSec int = 0
var client *tmt.Client

func init() {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = apiURL
	cpf.Language = "en-US"

	client, _ = tmt.NewClient(common.NewCredential(apiID, apiKey), "ap-shanghai", cpf)
}

/*
normal:
{
	"Response": {
		"RequestId": "f31036c2-b29d-40b9-b1f3-4309677fab1d",
		"Source": "en",
		"Target": "zh",
		"TargetText": "这是一次测试"
	}
}

error:
{
  "Response": {
    "RequestId": "ed93f3cb-f35e-473f-b9f3-0d451b8b79c6"
    "Error": {
      "Code": "AuthFailure.SignatureFailure",
      "Message": "The provided credentials could not be validated. Please check your signature is correct."
    },
  }
}
*/
type _Result struct {
	Response struct {
		RequestID  string `json:"RequestId"`
		Source     string `json:"Source"`
		Target     string `json:"Target"`
		TargetText string `json:"TargetText"`
		Error      struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"Response"`
}

// En2Cn english to chinese
func En2Cn(query string) (string, error) {
	return translate("en", "zh", query)
}

// Cn2En english to chinese
func Cn2En(query string) (string, error) {
	return translate("zh", "en", query)
}

func translate(from, to string, query string) (string, error) {
	start := time.Now()
	defer func() {
		timeCostMilliSec = int(time.Since(start) / (1000.0 * 1000.0))
	}()

	// if call cost less time than an interval, then sleep
	if timeCostMilliSec != 0 && timeCostMilliSec < interval {
		time.Sleep(time.Duration(interval-timeCostMilliSec+5) * time.Millisecond)
	}

	if len(query) >= LengthLimit {
		return "", errors.New("query lenth > limitation")
	}

	request := tmt.NewTextTranslateRequest()
	request.SourceText = common.StringPtr(query)
	request.Source = common.StringPtr(from)
	request.Target = common.StringPtr(to)
	request.ProjectId = common.Int64Ptr(0)

	resp, e := client.TextTranslate(request)
	if _, ok := e.(*terrors.TencentCloudSDKError); ok {
		return "", e // TODO wrap tencent error
	}

	if e != nil {
		return "", e
	}

	var rs _Result
	e = json.Unmarshal([]byte(resp.ToJsonString()), &rs)
	if e != nil {
		return "", e
	}

	return rs.Response.TargetText, nil
}
