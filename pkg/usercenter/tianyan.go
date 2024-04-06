package usercenter

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func tianUserOtpLogin(username, otp_code string) error {
	type auth struct {
		Username string `json:"username"`
		AuthCode string `json:"auth_code"`
	}

	url := "https://tianyan.bangdao-tech.com/apiv1/user/api/v1/login/auth"
	method := "POST"
	user := auth{
		Username: username,
		AuthCode: otp_code,
	}
	reqb, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, strings.NewReader(string(reqb)))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var rtn map[string]string
	err = json.Unmarshal(body, &rtn)
	if err != nil {
		return errors.Errorf("tianyan login unmarshal error %s", string(body))
	}
	if rtn["rtn_code"] == "000000" {
		return nil
	}
	return errors.New(string(body))
}
