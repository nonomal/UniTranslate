package logic

import (
	"context"
	"fmt"
	"uniTranslate/src/buffer"
	"uniTranslate/src/global"
	queueHandler "uniTranslate/src/service/queue/handler"
	"uniTranslate/src/service/web/handler"
	"uniTranslate/src/translate"
	"uniTranslate/src/types"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
)

// Translate 翻译
func Translate(ctx context.Context, ip string, req *types.TranslateReq) (data *types.TranslateData, err error) {
	textStr := gstr.Join(req.Text, "\n")
	// 内容转换为md5
	var keyStr string
	if global.CachePlatform {
		keyStr = fmt.Sprintf("to:%s-text:%s-platform:%s", req.To, textStr, req.Platform)
	} else {
		keyStr = fmt.Sprintf("to:%s-text:%s", req.To, textStr)
	}
	md5 := gmd5.MustEncrypt(keyStr)
	// 记录从翻译到获取到结果的时间
	startTime := gtime.Now().UnixMilli()
	// 创建所需要的参数
	translateReq := &translate.TranslateReq{
		HttpReq: &translate.TranslateHttpReq{
			ClientIp: ip,
			Context:  ctx,
		},
		From:     req.From,
		To:       req.To,
		Platform: req.Platform,
		Text:     req.Text,
		TextStr:  textStr,
	}
	// 判断是否进行缓存
	if global.CacheMode == "off" {
		data, err = translateHandler(translateReq)
	} else {
		dataT, err1 := global.GfCache.GetOrSetFunc(ctx, fmt.Sprintf("Translate:%s", md5), func(ctx context.Context) (value any, err error) {
			return translateHandler(translateReq)
		}, 0)
		if err1 != nil {
			err = err1
			return
		}
		if err = dataT.Scan(&data); err != nil {
			return
		}
	}
	nowTime := gtime.Now().UnixMilli()
	// 记录翻译
	queueHandler.RequestRecordQueue.Push(&types.RequestRecordData{
		ClientIp: ip,
		Body:     req,
		Time:     nowTime,
		Ok:       err == nil,
		ErrMsg:   err,
		Platform: data.Platform,
		TakeTime: int(nowTime - startTime),
		TraceId:  gtrace.GetTraceID(ctx),
	})
	return
}

// translateHandler 翻译处理
func translateHandler(req *translate.TranslateReq) (data *types.TranslateData, err error) {
	data, err = buffer.Buffer.Handler(req, handler.Translate)
	if data != nil {
		// 缓存写入数据库
		if global.CacheWriteToStorage {
			queueHandler.SaveQueue.Push(&types.SaveData{
				Data: data,
			})
		}
		// 翻译计数
		queueHandler.CountRecordQueue.Push(&types.CountRecordData{
			Data: data,
			Ok:   err == nil,
		})
	}
	return
}
