#!/usr/bin/env bash

export TELEGRAM_APITOKEN=""
export BOT_DEBUG="false"
export SUPER_USER_ID=""

export SMART_HOME_API_URL=""
export SMART_HOME_APITOKEN=""

export ALI_ASR_APPKEY=""
export ASR_API_URL=""
export AUDIO_SAMPLING_RATE_ASR="16000"

export ALI_ACCESS_KEYID=""
export ALI_ACCESS_KEYSECRET=""

export RedisAddress="localhost:6379"
export RedisDB=0
export RedisPassword=""

OS="$1"
if [[ "$OS" == "" ]]; then
  OS="osx"
fi

echo "run for $OS"
kill `pgrep -f go_telegram_bot_$OS`
./"go_telegram_bot_$OS"
