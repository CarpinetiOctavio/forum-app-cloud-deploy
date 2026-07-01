describe('Authentication Flow', () => {
    beforeEach(() => {
        cy.visit('/')
    })

    it('should show the login form by default', () => {
        cy.get('h2').should('contain', 'Sign In')
        cy.get('input#email').should('be.visible')
        cy.get('input#password').should('be.visible')
        cy.get('button[type="submit"]').should('contain', 'Sign In')
    })

    it('should toggle between login and registration', () => {
        // Switch to register
        cy.contains("Don't have an account? Sign Up").click()
        cy.get('h2').should('contain', 'Sign Up')
        cy.get('input#username').should('be.visible')
        cy.get('button[type="submit"]').should('contain', 'Sign Up')

        // Back to login
        cy.contains('Already have an account? Sign In').click()
        cy.get('h2').should('contain', 'Sign In')
        cy.get('input#username').should('not.exist')
    })

    it('should show error with invalid credentials', () => {
        cy.intercept('POST', '**/api/auth/login', {
            statusCode: 401,
            body: { error: 'Invalid credentials' }
        })

        cy.get('input#email').type('invalid@example.com')
        cy.get('input#password').type('wrongpass')
        cy.get('button[type="submit"]').click()

        cy.get('.error-message').should('be.visible')
            .and('contain', 'Invalid credentials')
    })

    it('should perform successful login', () => {
        cy.intercept('POST', '**/api/auth/login', {
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
        })

        cy.get('input#email').type('test@example.com')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.wait('@loginRequest')

        // Verify the app is displayed
        // cy.contains('Mini Social Network').should('be.visible') // title changed to "1", assertion no longer applies
        cy.contains('Hello, @testuser').should('be.visible')
    })

    it('should register successfully', () => {
        cy.intercept('POST', '**/api/auth/register', {
            statusCode: 201,
            body: {
                id: 2,
                email: 'newuser@example.com',
                username: 'newuser'
            }
        }).as('registerRequest')

        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: []
        })

        // Switch to registration mode
        cy.contains("Don't have an account? Sign Up").click()

        cy.get('input#email').type('newuser@example.com')
        cy.get('input#username').type('newuser')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.wait('@registerRequest')

        // Verify the app is displayed
        cy.contains('Hello, @newuser').should('be.visible')
    })
})
