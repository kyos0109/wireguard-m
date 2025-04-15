package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"github.com/kyos0109/WireGuard-M/models"
	"github.com/kyos0109/WireGuard-M/utils"
)

// GetPeers 回傳 JSON 格式的 peers 資料
func GetPeers(peerStore *models.PeerStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		i, ok := c.Params.Get("interfaceName")
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Found Interface args."})
			return
		}

		devicePeers, err := peerStore.LoadPeers()

		var filterPeers []models.Peer
		d := *devicePeers
		filterPeers = d[i]

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, filterPeers)
	}
}

func AddPeerPage(c *gin.Context) {
	c.HTML(http.StatusOK, "add_peer.html", nil)
}

// AddPeer 新增 peer，資料由 JSON 提交
func AddPeer(peerStore *models.PeerStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		devPeers := &models.DevicePeers{}
		if err := c.ShouldBindJSON(devPeers); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 先利用 wgctrl-go 更新 WireGuard 介面
		for dev, peers := range *devPeers {
			for _, peer := range peers {
				if err := utils.AddPeerToInterface(dev, &peer); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		// 儲存至 peers.json 檔案
		if err := peerStore.AddPeer(devPeers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Peer 已新增"})
	}
}

// DeletePeer 根據 POST 參數 id 刪除 peer
func DeletePeer(peerStore *models.PeerStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		interfaceName := c.PostForm("interface")
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 id"})
			return
		}
		peer, err := peerStore.GetPeerByID(interfaceName, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 從 WireGuard 介面移除該 peer
		if err := utils.RemovePeerFromInterface(interfaceName, peer.PublicKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 從 peers.json 刪除該 peer
		if err := peerStore.DeletePeer(interfaceName, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Peer 已刪除"})
	}
}

// GeneratePeerQR 根據路由參數 id 產生指定 peer 的 QR Code
func GeneratePeerQR(peerStore *models.PeerStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		i, ok := c.GetQuery("interface")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Interface Not Found."})
		}

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 id"})
			return
		}

		peer, err := peerStore.GetPeerByID(i, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config, err := utils.GeneratePeerConfig(peer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 產生 QR Code PNG 圖片
		png, err := qrcode.Encode(config, qrcode.Medium, 256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "image/png", png)
	}
}

func DownloadConfig(peerStore *models.PeerStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		i, ok := c.GetQuery("interface")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Interface Not Found."})
		}

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 id"})
			return
		}

		peer, err := peerStore.GetPeerByID(i, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config, err := utils.GeneratePeerConfig(peer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fileName := peer.Name + ".cfg"
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		c.Data(http.StatusOK, "application/octet-stream", []byte(config))
	}
}
