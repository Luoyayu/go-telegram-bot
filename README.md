[![Build Status](https://travis-ci.com/Luoyayu/go-telegram-bot.svg?branch=master)](https://travis-ci.com/Luoyayu/go-telegram-bot) [![Coverage Status](https://coveralls.io/repos/github/Luoyayu/go-telegram-bot/badge.svg?branch=master)](https://coveralls.io/github/Luoyayu/go-telegram-bot?branch=master)
 

# env variable

 export TELEGRAM_APITOKEN=""      
 export BOT_DEBUG="false"     
 export SUPER_USER_ID=""     
  export SUPER_USER_NAME=""       

> costumed service    

export SMART_HOME_API_URL=""    
export SMART_HOME_APITOKEN=""      

> Ali Yun ASR Service       
make sure you have **ffmpeg** in PATH to convert .oag(48k) to .wav(18k)   

export ALI_ASR_APPKEY=""    
export ALI_ACCESS_KEYID=""      
export ALI_ACCESS_KEYSECRET=""      
 
export ASR_API_URL="http://nls-gateway.cn-shanghai.aliyuncs.com/stream/v1/asr"     
export AUDIO_SAMPLING_RATE_ASR="16000" # 16000 or 8000      

> redis   

export RedisAddress="localhost:6379"    
export RedisDB=0   
export RedisPassword=""    
 
# Redis


AliToken is store in redis and set EXPIRE time by Gettoken API   


# build   

`bash build.sh [osx/windows/linux]`

# run  
please fill in the env variable in `run.sh`    
keep env blank if you don't need the service    

`bash run.sh [osx/windows/linux]`

# TODO

- [ ] [rsshub](https://docs.rsshub.app)   


# reference

[ALi ASR](https://nls-portal.console.aliyun.com/overview)    
[Ali GetToken](https://help.aliyun.com/document_detail/72153.html)     
[Telegram Bot API](https://core.telegram.org/api)      
[Telegram Bot API Methods](https://core.telegram.org/methods)     
[Telegram Bot Platform](https://telegram.org/blog/bot-revolution)    
[Golang Telegram Bot Api](https://github.com/go-telegram-bot-api/telegram-bot-api)    


# Licence

MIT   

