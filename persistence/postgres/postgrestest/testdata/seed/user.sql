/*
 Foo bar
 */
INSERT INTO platform_users (user_id, name, image_url)
VALUES ('user1', 'User One', 'https://example.com/user1.png'),
       ('user2', 'User Two', 'https://example.com/user2.png'),
       ('user3', 'User Three', NULL),
       ('user4', 'User Four', NULL);

-- name: insert-additional-users
INSERT INTO platform_users (user_id, name, image_url)
VALUES ('user5', 'User Five', NULL),
       ('user6', 'User Six', NULL);

