package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ThroughSuite struct {
	BaseTestSuite
}

func (s *ThroughSuite) insertFixtures() ([]*User, []*Post) {
	require := s.Require()

	var (
		userStore     = NewUserStore(s.db)
		postStore     = NewPostStore(s.db)
		userPostStore = NewUserPostStore(s.db)
	)

	users := []*User{
		NewUser("a"),
		NewUser("b"),
	}

	for _, u := range users {
		require.NoError(userStore.Debug().Insert(u))
	}

	posts := []*Post{
		NewPost("a"),
		NewPost("b"),
		NewPost("c"),
		NewPost("d"),
		NewPost("e"),
		NewPost("f"),
	}

	for _, p := range posts {
		require.NoError(postStore.Debug().Insert(p))
	}

	userPosts := []*UserPost{
		NewUserPost(users[0], posts[0]),
		NewUserPost(users[0], posts[1]),
		NewUserPost(users[0], posts[2]),
		NewUserPost(users[1], posts[3]),
		NewUserPost(users[1], posts[4]),
		NewUserPost(users[1], posts[5]),
	}

	for _, up := range userPosts {
		require.NoError(userPostStore.Debug().Insert(up))
	}

	return users, posts
}

func (s *ThroughSuite) TestFind() {
	s.insertFixtures()
	require := s.Require()

	q := NewUserQuery().
		WithPosts(nil, nil)
	users, err := NewUserStore(s.db).Debug().FindAll(q)
	require.NoError(err)

	require.Len(users, 2)
	require.Equal("a", users[0].Name)
	require.Len(users[0].Posts, 3)
	for i, txt := range []string{"a", "b", "c"} {
		require.Equal(txt, users[0].Posts[i].Text)
	}

	require.Equal("b", users[1].Name)
	require.Len(users[1].Posts, 3)
	for i, txt := range []string{"d", "e", "f"} {
		require.Equal(txt, users[1].Posts[i].Text)
	}
}

func (s *ThroughSuite) TestFind_Single() {
	s.insertFixtures()
	require := s.Require()

	q := NewPostQuery().
		WithUser(nil, nil)
	posts, err := NewPostStore(s.db).Debug().FindAll(q)
	require.NoError(err)

	require.Len(posts, 6)
	for i, p := range []string{"a", "b", "c", "d", "e", "f"} {
		require.Equal(p, posts[i].Text)
	}

	userByPost := map[string]string{
		"a": "a",
		"b": "a",
		"c": "a",
		"d": "b",
		"e": "b",
		"f": "b",
	}
	for _, p := range posts {
		require.Equal(userByPost[p.Text], p.User.Name)
	}
}

func TestThrough(t *testing.T) {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			id serial primary key,
			text text not null
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id serial primary key,
			name text not null
		)`,
		`CREATE TABLE IF NOT EXISTS user_posts (
			id serial primary key,
			user_id bigint not null references users(id),
			post_id bigint not null references posts(id)
		)`,
	}
	suite.Run(t, &ThroughSuite{NewBaseSuite(
		schema,
		"user_posts", "posts", "users",
	)})
}
