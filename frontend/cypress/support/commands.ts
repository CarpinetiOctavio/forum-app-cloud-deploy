export interface MockUser {
  id: number;
  email: string;
  username: string;
}

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Mocks a successful login for `user` and drives the real login form
       * to completion. Callers that need a specific GET /api/posts response
       * must register that intercept before calling this, since the app
       * fetches the post list immediately after login.
       */
      login(user: MockUser): Chainable<void>;
    }
  }
}

Cypress.Commands.add('login', (user: MockUser) => {
  cy.intercept('POST', '**/api/auth/login', {
    statusCode: 200,
    body: user,
  }).as('loginRequest');

  cy.visit('/');
  cy.get('input#email').type(user.email);
  cy.get('input#password').type('password123');
  cy.get('button[type="submit"]').click();
  cy.wait('@loginRequest');
  cy.contains(`Hello, @${user.username}`).should('be.visible');
});

export {};
