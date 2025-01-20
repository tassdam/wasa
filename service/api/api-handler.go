package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// My additions

	// "Login" part
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// "User" part
	rt.router.GET("/users", rt.wrap(rt.getMyPhoto))
	rt.router.PUT("/users/me", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/users/search", rt.wrap(rt.searchUsers))

	// "Conversations" part
	rt.router.GET("/users/me/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversation))

	// "Messages" part
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/conversations/:conversationId/messages/:messageId/forward", rt.wrap(rt.forwardMessage))

	// // "Comments" part
	// rt.router.POST("/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.commentMessage))
	// rt.router.DELETE("/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrap(rt.uncommentMessage))

	// // "Groups" part
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/users/me/groups", rt.wrap(rt.getMyGroups))
	// rt.router.POST("/groups/:groupId/members", rt.wrap(rt.addToGroup))
	// rt.router.DELETE("/groups/:groupId/members/me", rt.wrap(rt.leaveGroup))
	// rt.router.PUT("/groups/:groupId", rt.wrap(rt.setGroupName))
	// rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
