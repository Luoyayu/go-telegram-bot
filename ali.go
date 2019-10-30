package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	AliSDK "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	AliHttp "github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type getTokenFromSDKType struct {
	TokenType `json:"Token"`
	//RequestId    string `json:"RequestId"`
	//NlsRequestId string `json:"NlsRequestId"`
}

type TokenType struct {
	Id         string `json:"Id"` // token
	ExpireTime int64  `json:"ExpireTime"`
	UserId     string `json:"UserId"`
}

func getTokenFromSDK() (result *getTokenFromSDKType, contentString string) {
	client, err := AliSDK.NewClientWithAccessKey(
		"cn-shanghai",
		os.Getenv("ALI_ACCESS_KEYID"),
		os.Getenv("ALI_ACCESS_KEYSECRET"))
	if err != nil {
		Logger.Error(err)
	}
	request := AliHttp.NewCommonRequest()
	request.Method = "POST"
	request.Domain = "nls-meta.cn-shanghai.aliyuncs.com"
	request.ApiName = "CreateToken"
	request.Version = "2019-02-28"
	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		Logger.ErrorService("ali", err)
		return nil, err.Error()
	}
	contentString = response.GetHttpContentString()

	if response.GetHttpStatus() == 200 {
		log.Println(response.GetHttpContentString())
		err = json.Unmarshal(response.GetHttpContentBytes(), &result)
		if err != nil {
			Logger.ErrorService("ali", err)
		} else {
			Logger.InfofService("ali", "%+v\n", *result)
		}
		return result, contentString
	}
	return nil, contentString
}

func updateTokenAndStore() error {
	tokenFromSDK, contentString := getTokenFromSDK()
	if tokenFromSDK.Id == "" {
		return fmt.Errorf("try get token from sdk error: %s", contentString)
	} else {
		AliToken = tokenFromSDK.Id
		// store new token to redis and set it's expire time
		dbClient.Set("ALI_ACCESS_TOKEN", tokenFromSDK.Id,
			time.Duration(tokenFromSDK.ExpireTime*1e9))
	}
	return nil
}

func handleVoiceMsg2Text(
	appkey string,
	token string,
	fileName string,
	format string,
	sampleRate string,
	enablePunctuationPrediction bool,
	enableInverseTextNormalization bool,
	enableVoiceDetection bool) recognitionType {

	var asrUrl = os.Getenv("ASR_API_URL")
	asrUrl = asrUrl + "?appkey=" + appkey
	asrUrl = asrUrl + "&format=" + format
	asrUrl = asrUrl + "&sample_rate=" + sampleRate
	if enablePunctuationPrediction {
		asrUrl = asrUrl + "&enable_punctuation_prediction=" + "true"
	}
	if enableInverseTextNormalization {
		asrUrl = asrUrl + "&enable_inverse_text_normalization=" + "true"
	}
	if enableVoiceDetection {
		asrUrl = asrUrl + "&enable_voice_detection=" + "false"
	}
	Logger.Info(asrUrl)

	audioData, err := ioutil.ReadFile(fileName)
	if err != nil {
		Logger.Error(err)
	}
	request, err := http.NewRequest("POST", asrUrl, bytes.NewBuffer(audioData))
	if err != nil {
		Logger.Error(err)

	}

	request.Header.Add("X-NLS-Token", token)
	request.Header.Add("Content-Type", "application/octet-stream")

	var resultMap recognitionType
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		Logger.Error(err)
		resultMap.Error = err
		return resultMap
	} else {
		defer response.Body.Close()
	}

	statusCode := response.StatusCode

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&resultMap)

	if err != nil {
		Logger.Error(err)
	}

	if statusCode == 200 {
		var result = resultMap.Result
		Logger.Infof("recognition succeed, result is %q", result)
	} else {
		Logger.Error("recognition failed，HTTP StatusCode: " + strconv.Itoa(statusCode))
	}
	return resultMap
}

type recognitionType struct {
	TaskId  string `json:"task_id"`
	Result  string `json:"result"`
	Status  int64  `json:"status"`
	Message string `json:"message"`
	Error   error
}
