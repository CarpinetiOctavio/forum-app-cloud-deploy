// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

// Comando personalizado para login
Cypress.Commands.add('login', (email, password) => {
    cy.visit('/')
    cy.get('input[type="email"]').type(email)
    cy.get('input[type="password"]').type(password)
    cy.get('button[type="submit"]').click()
})

// Comando para interceptar API del backend
Cypress.Commands.add('mockBackend', () => {
    cy.intercept('POST', '**/api/login', {
        statusCode: 200,
        body: {
            id: 1,
            email: 'test@example.com',
            username: 'testuser'
        }
    }).as('loginRequest')

    cy.intercept('GET', '**/api/posts', {
        statusCode: 200,
        body: []
    }).as('getPosts')
})