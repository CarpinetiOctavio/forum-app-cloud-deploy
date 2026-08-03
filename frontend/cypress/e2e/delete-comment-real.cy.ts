// Non-mocked by design (ADR-003 / ADR-005): DeleteComment's authorization
// check lives in the repository's SQL `WHERE ... AND user_id = ?`, not in
// the service layer. A mocked test can't reach that clause -- this one runs
// against the real backend and a real SQLite database. No cy.intercept()
// anywhere in this file.
//
// Requires the real backend to be running at localhost:8080 (see the
// cypress-e2e CI job, or run `go run ./cmd/api` locally before `cypress run`).

const API = 'http://localhost:8080/api';
const runId = Date.now();

const postAuthor = {
  email: `post-author-${runId}@test.com`,
  username: `postauthor${runId}`,
  password: 'password123',
};

const commentAuthor = {
  email: `comment-author-${runId}@test.com`,
  username: `commentauthor${runId}`,
  password: 'password123',
};

const postTitle = `Real backend test post ${runId}`;
const commentContent = `A real comment, authored by commentAuthor ${runId}`;

let postId: number;
let commentId: number;
let postAuthorId: number;
let commentAuthorId: number;

describe('Delete Comment -- real backend, no mocks', () => {
  before(() => {
    // Seed via real HTTP requests against the real backend (cy.request, not
    // cy.intercept) -- exercises the same registration/creation path a real
    // user would, instead of reaching around it into SQLite directly.
    cy.request('POST', `${API}/auth/register`, postAuthor).then((res) => {
      postAuthorId = res.body.id;
    });

    cy.request('POST', `${API}/auth/register`, commentAuthor).then((res) => {
      commentAuthorId = res.body.id;
    });

    cy.then(() => {
      cy.request({
        method: 'POST',
        url: `${API}/posts`,
        headers: { 'X-User-ID': String(postAuthorId) },
        body: { title: postTitle, content: 'Content for the real DeleteComment test' },
      }).then((res) => {
        postId = res.body.id;
      });
    });

    cy.then(() => {
      cy.request({
        method: 'POST',
        url: `${API}/posts/${postId}/comments`,
        headers: { 'X-User-ID': String(commentAuthorId) },
        body: { content: commentContent },
      }).then((res) => {
        commentId = res.body.id;
      });
    });
  });

  const loginAs = (user: typeof postAuthor) => {
    cy.visit('/');
    cy.get('input#email').type(user.email);
    cy.get('input#password').type(user.password);
    cy.get('button[type="submit"]').click();
    cy.contains(`Hello, @${user.username}`).should('be.visible');
  };

  const goToPost = () => {
    cy.contains(postTitle).click();
    cy.contains('← Back').should('be.visible');
  };

  it("rejects deletion by a user who is not the comment's author, even via a direct API call", () => {
    loginAs(postAuthor);
    goToPost();
    cy.contains(commentContent).should('be.visible');

    // The UI already hides the button for a non-author -- confirmed first,
    // but that alone doesn't prove the server rejects the request too.
    cy.get('.comment-delete-btn').should('not.exist');

    // Same request the frontend would send, sent directly: proves the real
    // repository WHERE clause rejects it, not just that the button is
    // hidden from someone using the app normally.
    cy.request({
      method: 'DELETE',
      url: `${API}/posts/${postId}/comments/${commentId}`,
      headers: { 'X-User-ID': String(postAuthorId) },
      failOnStatusCode: false,
    }).then((res) => {
      expect(res.status).to.equal(403);
      expect(res.body.error).to.equal('you do not have permission to delete this comment or it does not exist');
    });

    // Still there -- the rejected request didn't remove it.
    cy.request('GET', `${API}/posts/${postId}/comments`).then((res) => {
      expect(res.body.map((c: { id: number }) => c.id)).to.include(commentId);
    });
  });

  it('allows deletion by the real comment author, through the real UI', () => {
    loginAs(commentAuthor);
    goToPost();
    cy.contains(commentContent).should('be.visible');

    cy.get('.comment-delete-btn').click();

    cy.contains('Comment deleted successfully').should('be.visible');
    cy.contains(commentContent).should('not.exist');

    cy.request('GET', `${API}/posts/${postId}/comments`).then((res) => {
      expect(res.body.map((c: { id: number }) => c.id)).to.not.include(commentId);
    });
  });
});
