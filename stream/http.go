package main

import (
	"log"
	"time"

	"github.com/deepch/vdk/av"
	webrtc "github.com/deepch/vdk/format/webrtcv3"
	"github.com/gin-gonic/gin"
)

type JCodec struct {
	Type string
}

type WebRTCInfo struct {
	Codecs []JCodec `json:"codecs"`
	Answer string   `json:"answer"`
}

func serveHTTP() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/stream/webrtc/:uuid", HTTPAPIServerStreamWebRTC)
	log.Println("http server ", Config.Server.HTTPPort)
	err := router.Run(Config.Server.HTTPPort)
	if err != nil {
		log.Fatalln("http server error:", err)
	}
}

func HTTPAPIServerStreamWebRTC(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	suuid := c.Param("uuid")
	
	// Check if stream exists
	if !Config.ext(suuid) {
		c.JSON(404, gin.H{"error": "stream not found"})
		return
	}
	
	// Start stream if needed
	Config.RunIFNotRun(suuid)
	
	// Get codecs
	codecs := Config.coGe(suuid)
	if codecs == nil {
		c.JSON(500, gin.H{"error": "codec not found"})
		return
	}
	
	// Create codec info
	var tmpCodec []JCodec
	var AudioOnly bool
	if len(codecs) == 1 && codecs[0].Type().IsAudio() {
		AudioOnly = true
	}
	
	for _, codec := range codecs {
		if codec.Type() != av.H264 && codec.Type() != av.PCM_ALAW && codec.Type() != av.PCM_MULAW && codec.Type() != av.OPUS {
			log.Println("codec not supported:", codec.Type())
			continue
		}
		if codec.Type().IsVideo() {
			tmpCodec = append(tmpCodec, JCodec{Type: "video"})
		} else {
			tmpCodec = append(tmpCodec, JCodec{Type: "audio"})
		}
	}
	
	// Get WebRTC offer from request
	offer := c.PostForm("data")
	if offer == "" {
		c.JSON(400, gin.H{"error": "missing WebRTC offer"})
		return
	}
	
	// Create WebRTC connection
	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{
		ICEServers: Config.GetICEServers(), 
		PortMin: Config.GetWebRTCPortMin(), 
		PortMax: Config.GetWebRTCPortMax(),
	})
	
	answer, err := muxerWebRTC.WriteHeader(codecs, offer)
	if err != nil {
		c.JSON(500, gin.H{"error": "WebRTC header error: " + err.Error()})
		return
	}
	
	// Start streaming goroutine
	go func() {
		cid, ch := Config.clAd(suuid)
		defer Config.clDe(suuid, cid)
		defer muxerWebRTC.Close()
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				log.Println("novideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame || AudioOnly {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart && !AudioOnly {
					continue
				}
				err = muxerWebRTC.WritePacket(pck)
				if err != nil {
					log.Println("packet:", err)
					return
				}
			}
		}
	}()
	
	// Return combined response - note that we keep the answer as base64
	// The client will decode it using atob()
	c.JSON(200, WebRTCInfo{
		Codecs: tmpCodec,
		Answer: answer, // This is already base64 encoded by the WriteHeader function
	})
}