package routes

import (
	"myapp/controllers"
	"myapp/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate2()) //only if user has a token can he access these routes
	incomingRoutes.GET("/user", controllers.GetOwnUser())
	incomingRoutes.DELETE("/user/delete", controllers.DeleteUser())
	incomingRoutes.PUT("/user/update", controllers.UpdateUser())
	incomingRoutes.POST("/user/logout", controllers.Logout())
	incomingRoutes.POST("/user/follow", controllers.FollowAnotherUser())
	incomingRoutes.POST("/user/avatarupdate", controllers.UploadAvatar())
	incomingRoutes.POST("/user/makepost", controllers.MakeAPost())
	incomingRoutes.DELETE("/user/deletepost", controllers.DeleteAPost())
	incomingRoutes.PUT("/user/updatepost/:post_id", controllers.UpdatePost())
	incomingRoutes.GET("/user/readpost", controllers.ReadAPost())
	incomingRoutes.GET("/user/readallpostofauser", controllers.ReadAllPostOfAUser())
	incomingRoutes.POST("/user/likepost", controllers.LikeAPost())
	incomingRoutes.POST("/user/commentonpost/:post_id", controllers.CommentOnPost())
	incomingRoutes.GET("/user/searchbytags", controllers.SearchPostsByTags())
	incomingRoutes.POST("/user/bookmarkpost", controllers.BookmarkPost())
}
