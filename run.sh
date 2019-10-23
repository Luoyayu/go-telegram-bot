#!/usr/bin/env bash

export TELEGRAM_APITOKEN=""
export GRANTEDIDS=","
export BOT_DEBUG="false"
export CHAT_ID=""

export SMART_HOME_API_URL=""
export SMART_HOME_APITOKEN=""

export ALI_ASR_APPKEY=""
export ALI_ACCESS_TOKEN=""
export ASR_API_URL="http://nls-gateway.cn-shanghai.aliyuncs.com/stream/v1/asr"
export AUDIO_SAMPLING_RATE_ASR="16000"

export ALI_ACCESS_KEYID=""
export ALI_ACCESS_KEYSECRET=""

./go_telegram_bot
