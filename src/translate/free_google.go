package translate

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"regexp"
)

type FreeGoogle struct {
}

func (f FreeGoogle) Translate(request *TranslateReq) (resp []*TranslateResp, err error) {
	var responses []*TranslateResp
	fmt.Println("测试", request)
	for _, str := range request.Text {
		response, err1 := translatedContent(str, request.From, request.To)

		if err1 != nil {
			err = fmt.Errorf("翻译错误: %s", err.Error())
			return
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (f FreeGoogle) GetMode() (mode string) {
	return FreeGoogleMode
}

func translatedContent(text, fromLanguage string, targetLanguage string) (*TranslateResp, error) {
	url := "https://translate.google.com/_/TranslateWebserverUi/data/batchexecute?rpcids=MkEWBc&f.sid=-2609060161424095358&bl=boq_translate-webserver_20201203.07_p0&hl=zh-CN&soc-app=1&soc-platform=1&soc-device=1&_reqid=359373&rt=c"
	headers := map[string]string{
		"Accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36",
	}
	post, err := g.Client().SetHeaderMap(headers).Post(gctx.New(), url, g.MapStrStr{
		"f.req": fmt.Sprintf(`[[["MkEWBc","[[\"%s\",\"%s\",\"%s\",true],[null]]",null,"generic"]]]`, text, fromLanguage, targetLanguage),
	})
	if err != nil {
		return nil, err
	}
	bodyString := post.ReadAllString()
	re := regexp.MustCompile(`,\[\[\\"(.*?)\\",\[\\`)
	matches := re.FindStringSubmatch(bodyString)
	if len(matches) > 1 {
		str := matches[1]
		from := gstr.Split(str, `,\"`)
		fromLen := len(from)

		return &TranslateResp{
			Text:     gstr.Split(str, `\",`)[0],
			FromLang: from[fromLen-1],
		}, nil
	}
	return nil, fmt.Errorf("translate error")
}
