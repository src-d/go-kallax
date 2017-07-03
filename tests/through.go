package tests

import kallax "gopkg.in/src-d/go-kallax.v1"

type User struct {
	kallax.Model `table:"users" pk:"id,autoincr"`
	ID           int64
	Name         string
	Posts        []*Post `through:"user_posts"`
}

func newUser(name string) *User {
	return &User{Name: name}
}

type Post struct {
	kallax.Model `table:"posts" pk:"id,autoincr"`
	ID           int64
	Text         string
	User         *User `through:"user_posts"`
}

func newPost(text string) *Post {
	return &Post{Text: text}
}

type UserPost struct {
	kallax.Model `table:"user_posts" pk:"id,autoincr"`
	ID           int64
	User         *User `fk:",inverse"`
	Post         *Post `fk:",inverse"`
}

func newUserPost(user *User, post *Post) *UserPost {
	return &UserPost{User: user, Post: post}
}
