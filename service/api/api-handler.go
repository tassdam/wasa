package api

import (
	"net/http"
)

func (rt *_router) Handler() http.Handler {
	rt.router.POST("/session", rt.wrap(rt.doLogin))
	rt.router.GET("/users/photo", rt.wrap(rt.getMyPhoto))
	rt.router.PUT("/users/photo", rt.wrap(rt.setMyPhoto))
	rt.router.PUT("/users/name", rt.wrap(rt.setMyUserName))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/groups", rt.wrap(rt.getMyGroups))
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/search", rt.wrap(rt.searchUsers))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversation))
	rt.router.POST("/conversations/:conversationId/message", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/message/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/conversations/:conversationId/message/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/conversations/:conversationId/message/:messageId/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationId/message/:messageId/comment", rt.wrap(rt.uncommentMessage))
	rt.router.GET("/groups/:groupId", rt.wrap(rt.getGroup))
	rt.router.DELETE("/groups/:groupId", rt.wrap(rt.leaveGroup))
	rt.router.POST("/groups/:groupId", rt.wrap(rt.addToGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.GET("/liveness", rt.liveness)
	return rt.router
}
