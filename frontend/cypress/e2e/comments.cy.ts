import type { MockUser } from '../support/commands';

const currentUser: MockUser = { id: 1, email: 'test@example.com', username: 'testuser' };

const post = {
  id: 1,
  title: 'Test post',
  content: 'Post content',
  user_id: currentUser.id,
  username: currentUser.username,
  created_at: '2026-01-01T10:00:00Z',
};

const ownComment = {
  id: 1,
  post_id: post.id,
  user_id: currentUser.id,
  username: currentUser.username,
  content: 'Original comment',
  created_at: '2026-01-01T11:00:00Z',
};

const otherUsersComment = {
  id: 2,
  post_id: post.id,
  user_id: 999,
  username: 'otheruser',
  content: "Someone else's comment",
  created_at: '2026-01-01T11:00:00Z',
};

const visitPostDetail = () => {
  cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [post] });
  cy.login(currentUser);
  cy.intercept('GET', '**/api/posts/1', { statusCode: 200, body: post });
  cy.contains(post.title).click();
  cy.contains('← Back').should('be.visible');
};

describe('Create Comment', () => {
  it('creates a comment successfully', () => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [] }).as('getComments');
    visitPostDetail();
    cy.wait('@getComments');

    cy.intercept('POST', '**/api/posts/1/comments', {
      statusCode: 201,
      body: ownComment,
    }).as('createComment');
    // The comment list re-fetches after creation.
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [ownComment] });

    cy.get('textarea[placeholder*="comment"]').type(ownComment.content);
    cy.contains('button', 'Comment').click();

    cy.wait('@createComment');
    cy.contains(ownComment.content).should('be.visible');
  });
});

describe('Validation', () => {
  it('disables the comment button when the content is empty', () => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [] });
    visitPostDetail();

    cy.contains('button', 'Comment').should('be.disabled');
    cy.get('textarea[placeholder*="comment"]').type('Something');
    cy.contains('button', 'Comment').should('not.be.disabled');
  });
});

describe('Edit Comment', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [ownComment] });
    visitPostDetail();
    cy.contains(ownComment.content).should('be.visible');
  });

  it('edits a comment successfully', () => {
    const updated = { ...ownComment, content: 'Updated comment' };
    cy.intercept('PUT', '**/api/posts/1/comments/1', { statusCode: 200, body: updated }).as('editComment');

    cy.get('.comment-edit-btn').click();
    cy.get('.comment-edit-form textarea').should('have.value', ownComment.content).clear().type(updated.content);
    cy.get('.comment-save-btn').click();

    cy.wait('@editComment');
    cy.contains(updated.content).should('be.visible');
    cy.get('.comment-edit-form').should('not.exist');
  });

  it('cancels editing without saving', () => {
    cy.get('.comment-edit-btn').click();
    cy.get('.comment-edit-form textarea').clear().type('Discarded content');
    cy.get('.comment-cancel-btn').click();

    cy.get('.comment-edit-form').should('not.exist');
    cy.contains(ownComment.content).should('be.visible');
  });

  it('shows an error and stays in edit mode when the edit is rejected', () => {
    cy.intercept('PUT', '**/api/posts/1/comments/1', {
      statusCode: 403,
      body: { error: 'you do not have permission to edit this comment' },
    }).as('editComment');

    cy.get('.comment-edit-btn').click();
    cy.get('.comment-save-btn').click();

    cy.wait('@editComment');
    cy.contains('you do not have permission to edit this comment').should('be.visible');
    cy.get('.comment-edit-form').should('exist');
  });
});

describe("Another user's comment", () => {
  it('does not show the edit button', () => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [otherUsersComment] });
    visitPostDetail();

    cy.contains(otherUsersComment.content).should('be.visible');
    cy.get('.comment-edit-btn').should('not.exist');
  });

  it('does not show the delete button', () => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [otherUsersComment] });
    visitPostDetail();

    cy.contains(otherUsersComment.content).should('be.visible');
    cy.get('.comment-delete-btn').should('not.exist');
  });
});

describe('Delete Comment', () => {
  it('deletes a comment successfully', () => {
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [ownComment] });
    visitPostDetail();
    cy.contains(ownComment.content).should('be.visible');

    cy.intercept('DELETE', '**/api/posts/1/comments/1', { statusCode: 200, body: { message: 'comment deleted' } }).as(
      'deleteComment'
    );

    cy.get('.comment-delete-btn').click();

    cy.wait('@deleteComment');
    cy.contains(ownComment.content).should('not.exist');
    cy.contains('Comment deleted successfully').should('be.visible');
  });
});
