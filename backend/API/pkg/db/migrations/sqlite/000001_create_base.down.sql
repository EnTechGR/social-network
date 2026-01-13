DROP TABLE IF EXISTS notifications;
DROP INDEX IF EXISTS idx_notifications_type;
DROP INDEX IF EXISTS idx_notifications_user;

DROP TABLE IF EXISTS group_message_images;
DROP TABLE IF EXISTS group_messages;

DROP TABLE IF EXISTS chat_images;
DROP INDEX IF EXISTS idx_chat_images_message;

DROP TABLE IF EXISTS messages;
DROP INDEX IF EXISTS idx_messages_created;
DROP INDEX IF EXISTS idx_messages_receiver;
DROP INDEX IF EXISTS idx_messages_sender;

DROP TABLE IF EXISTS follow_relationships;
DROP INDEX IF EXISTS idx_follower_status;
DROP INDEX IF EXISTS idx_followee_status;

DROP TABLE IF EXISTS reactions;
DROP INDEX IF EXISTS idx_reactions_comment;
DROP INDEX IF EXISTS idx_reactions_post;

DROP TABLE IF EXISTS comments;
DROP INDEX IF EXISTS idx_comments_user;
DROP INDEX IF EXISTS idx_comments_post;

DROP TABLE IF EXISTS images;
DROP INDEX IF EXISTS idx_images_post;

DROP TABLE IF EXISTS post_allowed_users;
DROP INDEX IF EXISTS idx_post_allowed_user;

DROP TABLE IF EXISTS post_categories;
DROP INDEX IF EXISTS idx_post_categories_cat;
DROP INDEX IF EXISTS idx_post_categories_post;

DROP TABLE IF EXISTS posts;
DROP INDEX IF EXISTS idx_posts_group;
DROP INDEX IF EXISTS idx_posts_visibility;
DROP INDEX IF EXISTS idx_posts_user;

DROP TABLE IF EXISTS group_event_votes;
DROP TABLE IF EXISTS group_event_options;
DROP TABLE IF EXISTS group_events;
DROP INDEX IF EXISTS idx_group_events_group;

DROP TABLE IF EXISTS group_join_requests;
DROP TABLE IF EXISTS group_invites;
DROP TABLE IF EXISTS group_members;
DROP INDEX IF EXISTS idx_group_members_user;
DROP INDEX IF EXISTS idx_group_members_status;

DROP TABLE IF EXISTS groups;
DROP INDEX IF EXISTS idx_groups_public;

DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS oauth_states;
DROP INDEX IF EXISTS idx_oauth_states_expires;

DROP TABLE IF EXISTS oauth_accounts;
DROP INDEX IF EXISTS idx_oauth_user_id;
DROP INDEX IF EXISTS idx_oauth_provider_user;

DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS idx_sessions_user;

DROP TABLE IF EXISTS user_auth;
DROP TABLE IF EXISTS user;
