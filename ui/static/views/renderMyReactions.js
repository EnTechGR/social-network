import { renderPosts } from "../views/components/posts.js";
import { navigateTo } from "../router.js";

export async function renderMyReactions(main) {
  // Render page structure
  main.innerHTML = `
    <section class="my-reactions-page fade-in">
    <header class="my-reactions-header">
        <h1 class="my-reactions-title">My Reactions</h1>
        <p class="my-reactions-description">
        A summary of the posts you’ve enjoyed and discussed across BookTalk.
        </p>
    </header>

    <div class="reactions-section">
        <h3>Liked Posts</h3>
        <div id="likedPostsContainer" class="forum-container"></div>
    </div>

    <div class="reactions-section">
        <h3>Disliked Posts</h3>
        <div id="dislikedPostsContainer" class="forum-container"></div>
    </div>
    </section>
  `;

  const likedContainer = document.getElementById("likedPostsContainer");
  const dislikedContainer = document.getElementById("dislikedPostsContainer");

  try {
    // Load both reactions in parallel
    const [liked, disliked] = await Promise.all([
      fetchReactionPosts("liked"),
      fetchReactionPosts("disliked"),
    ]);

    if (liked.length) renderPosts(likedContainer, liked, "/user");
    else
      likedContainer.innerHTML = `<p class="empty-feed">You haven’t liked any posts yet.</p>`;

    if (disliked.length) renderPosts(dislikedContainer, disliked, "/user");
    else
      dislikedContainer.innerHTML = `<p class="empty-feed">You haven’t disliked any posts yet.</p>`;
  } catch (err) {
    console.error("Error loading reactions:", err);
    likedContainer.innerHTML = `<p class="error-message">⚠️ Unable to load liked posts.</p>`;
    dislikedContainer.innerHTML = `<p class="error-message">⚠️ Unable to load disliked posts.</p>`;
  }
}

/* --------------------------
   FETCH POSTS BY REACTION TYPE
-------------------------- */
async function fetchReactionPosts(type) {
  try {
    const resp = await fetch(`http://localhost:8080/forum/api/user/${type}`, {
      credentials: "include",
    });
    if (!resp.ok) throw new Error(`Failed to load ${type} posts`);
    return await resp.json();
  } catch (err) {
    console.error(`Error fetching ${type} posts:`, err);
    return [];
  }
}
