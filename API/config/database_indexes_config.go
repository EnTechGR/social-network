package config

const IdxPostsUserID = `CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`
const IdxPostCategoriesPostID = `CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON post_categories(post_id);`
const IdxPostCategoriesCategoryID = `CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);`
const IdxCommentsPostID = `CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`
const IdxCommentsUserID = `CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);`
const IdxReactionsUserID = `CREATE INDEX IF NOT EXISTS idx_reactions_user_id ON reactions(user_id);`
const IdxReactionsPostID = `CREATE INDEX IF NOT EXISTS idx_reactions_post_id ON reactions(post_id);`
const IdxReactionsCommentID = `CREATE INDEX IF NOT EXISTS idx_reactions_comment_id ON reactions(comment_id);`
const IdxImagesPostID = `CREATE INDEX IF NOT EXISTS idx_images_post_id ON images(post_id);`
const IdxNotificationsUserID = `CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);`
const IdxNotificationsFromUserID = `CREATE INDEX IF NOT EXISTS idx_notifications_from_user_id ON notifications(from_user_id);`

// -- Index for faster lookups by provider and provider_user_id
const CreateOAuthIndexes = `
		CREATE INDEX IF NOT EXISTS idx_oauth_provider_user 
		ON oauth_accounts(provider, provider_user_id);

		CREATE INDEX IF NOT EXISTS idx_oauth_user_id 
		ON oauth_accounts(user_id);
		`

		// Message indexes for efficient queries
const IdxMessagesSenderID = `CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);`
const IdxMessagesReceiverID = `CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages(receiver_id);`
const IdxMessagesCreatedAt = `CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);`

// Composite index for conversation queries (sender + receiver + created_at)
const IdxMessagesConversation = `CREATE INDEX IF NOT EXISTS idx_messages_conversation 
    ON messages(sender_id, receiver_id, created_at DESC);`

// Index for unread messages queries
const IdxMessagesUnread = `CREATE INDEX IF NOT EXISTS idx_messages_unread 
    ON messages(receiver_id, is_read, created_at DESC);`