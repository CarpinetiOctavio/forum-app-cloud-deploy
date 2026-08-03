import type { MockUser } from '../support/commands';

const currentUser: MockUser = { id: 1, email: 'test@example.com', username: 'testuser' };

const ownPost = {
  id: 1,
  title: 'My first post',
  content: 'Original content',
  user_id: currentUser.id,
  username: currentUser.username,
  created_at: '2026-01-01T10:00:00Z',
};

const otherUsersPost = {
  id: 2,
  title: "Someone else's post",
  content: 'Not mine',
  user_id: 999,
  username: 'otheruser',
  created_at: '2026-01-01T10:00:00Z',
};

describe('Create Post', () => {
  it('creates a post successfully', () => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [] });
    cy.login(currentUser);

    cy.intercept('POST', '**/api/posts', {
      statusCode: 201,
      body: ownPost,
    }).as('createPost');

    // The list re-fetches after creation; this is what it returns then.
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [ownPost] });

    cy.get('input[placeholder*="title"]').type(ownPost.title);
    cy.get('textarea[placeholder*="share"]').type(ownPost.content);
    cy.contains('button', 'Publish Post').click();

    cy.wait('@createPost');
    cy.contains(ownPost.title).should('be.visible');
    cy.contains(ownPost.content).should('be.visible');
  });
});

describe('Validation', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [] });
    cy.login(currentUser);
  });

  it('does not submit when the title is empty', () => {
    cy.get('textarea[placeholder*="share"]').type('Content only');
    cy.contains('button', 'Publish Post').click();

    cy.get('input[placeholder*="title"]')
      .should('have.prop', 'validity')
      .and('have.property', 'valueMissing', true);
  });

  it('does not submit when the content is empty', () => {
    cy.get('input[placeholder*="title"]').type('Title only');
    cy.contains('button', 'Publish Post').click();

    cy.get('textarea[placeholder*="share"]')
      .should('have.prop', 'validity')
      .and('have.property', 'valueMissing', true);
  });

  it('shows the server error when the title is below the 3-character minimum', () => {
    // HTML5 required is satisfied (non-empty) but the service's own
    // minimum-length rule isn't -- this exercises the server-rejection
    // path, not the client-side required check the other two tests cover.
    cy.intercept('POST', '**/api/posts', {
      statusCode: 400,
      body: { error: 'title must be at least 3 characters' },
    }).as('createPost');

    cy.get('input[placeholder*="title"]').type('ab');
    cy.get('textarea[placeholder*="share"]').type('Some content');
    cy.contains('button', 'Publish Post').click();

    cy.wait('@createPost');
    cy.contains('title must be at least 3 characters').should('be.visible');
  });
});

describe('Edit Post', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [ownPost] });
    cy.login(currentUser);

    cy.intercept('GET', '**/api/posts/1', { statusCode: 200, body: ownPost });
    cy.intercept('GET', '**/api/posts/1/comments', { statusCode: 200, body: [] });
    cy.contains(ownPost.title).click();
    cy.contains('← Back').should('be.visible');
  });

  it('edits a post successfully', () => {
    const updated = { ...ownPost, title: 'Updated title', content: 'Updated content' };
    cy.intercept('PUT', '**/api/posts/1', { statusCode: 200, body: updated }).as('editPost');

    cy.contains('button', 'Edit').click();
    cy.get('.post-detail-edit-form input').should('have.value', ownPost.title).clear().type(updated.title);
    cy.get('.post-detail-edit-form textarea').should('have.value', ownPost.content).clear().type(updated.content);
    cy.contains('button', 'Save').click();

    cy.wait('@editPost');
    cy.contains(updated.title).should('be.visible');
    cy.contains(updated.content).should('be.visible');
    cy.get('.post-detail-edit-form').should('not.exist');
  });

  it('cancels editing without saving', () => {
    cy.contains('button', 'Edit').click();
    cy.get('.post-detail-edit-form input').clear().type('Discarded title');
    cy.contains('button', 'Cancel').click();

    cy.get('.post-detail-edit-form').should('not.exist');
    cy.contains(ownPost.title).should('be.visible');
  });

  it('shows an error and stays in edit mode when the edit is rejected', () => {
    cy.intercept('PUT', '**/api/posts/1', {
      statusCode: 403,
      body: { error: 'you do not have permission to edit this post' },
    }).as('editPost');

    cy.contains('button', 'Edit').click();
    cy.contains('button', 'Save').click();

    cy.wait('@editPost');
    cy.contains('you do not have permission to edit this post').should('be.visible');
    cy.get('.post-detail-edit-form').should('exist');
  });
});

describe("Another user's post", () => {
  it('does not show the delete button on the post list', () => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [otherUsersPost] });
    cy.login(currentUser);

    cy.contains(otherUsersPost.title).should('be.visible');
    cy.get('.post-card').within(() => {
      cy.contains('Delete').should('not.exist');
    });
  });

  it('does not show the edit button on the post detail', () => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [otherUsersPost] });
    cy.login(currentUser);

    cy.intercept('GET', '**/api/posts/2', { statusCode: 200, body: otherUsersPost });
    cy.intercept('GET', '**/api/posts/2/comments', { statusCode: 200, body: [] });
    cy.contains(otherUsersPost.title).click();

    cy.contains('← Back').should('be.visible');
    cy.contains('button', 'Edit').should('not.exist');
  });
});

describe('Delete Post', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [ownPost] });
    cy.login(currentUser);
  });

  it('deletes a post successfully', () => {
    cy.intercept('DELETE', '**/api/posts/1', { statusCode: 200, body: { message: 'post deleted' } }).as(
      'deletePost'
    );
    // The list re-fetches after deletion.
    cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [] });

    cy.contains('button', 'Delete').click();

    cy.wait('@deletePost');
    cy.contains(ownPost.title).should('not.exist');
    cy.contains('No posts yet').should('be.visible');
  });

  it('does not delete a post when the confirmation dialog is cancelled', () => {
    cy.on('window:confirm', () => false);
    cy.intercept('DELETE', '**/api/posts/1').as('deletePost');

    cy.contains('button', 'Delete').click();

    cy.get('@deletePost.all').should('have.length', 0);
    cy.contains(ownPost.title).should('be.visible');
  });
});
