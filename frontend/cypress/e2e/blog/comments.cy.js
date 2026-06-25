describe('Comments Management', () => {
    beforeEach(() => {
        // Login
        cy.visit('/')

        cy.intercept('POST', '**/api/auth/login', {
            statusCode: 200,
            body: { id: 1, email: 'test@example.com', username: 'testuser' }
        })

        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: [{
                id: 1,
                title: 'Post de prueba',
                content: 'Contenido del post',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }]
        })

        cy.get('input#email').type('test@example.com')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()
    })

    it('should show post detail on click', () => {
        cy.intercept('GET', '**/api/posts/1', {
            statusCode: 200,
            body: {
                id: 1,
                title: 'Post de prueba',
                content: 'Contenido del post',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }
        })

        cy.intercept('GET', '**/api/posts/1/comments', {
            statusCode: 200,
            body: []
        })

        cy.contains('Post de prueba').click()

        cy.contains('← Back').should('be.visible')
        cy.contains('Add Comment').should('be.visible')
    })

    it('should create a comment', () => {
        cy.intercept('GET', '**/api/posts/1', {
            statusCode: 200,
            body: {
                id: 1,
                title: 'Post de prueba',
                content: 'Contenido',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }
        })

        cy.intercept('GET', '**/api/posts/1/comments', {
            statusCode: 200,
            body: []
        }).as('getComments')

        cy.intercept('POST', '**/api/posts/1/comments', {
            statusCode: 201,
            body: {
                id: 1,
                post_id: 1,
                user_id: 1,
                username: 'testuser',
                content: 'Mi comentario',
                created_at: new Date().toISOString()
            }
        }).as('createComment')

        cy.contains('Post de prueba').click()
        cy.wait('@getComments')

        cy.get('textarea[placeholder*="comment"]').type('Mi comentario')
        cy.contains('button', 'Comment').click()

        cy.wait('@createComment')
    })

    it('should disable comment button when empty', () => {
        cy.intercept('GET', '**/api/posts/1', {
            statusCode: 200,
            body: {
                id: 1,
                title: 'Post de prueba',
                content: 'Contenido',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }
        })

        cy.intercept('GET', '**/api/posts/1/comments', {
            statusCode: 200,
            body: []
        })

        cy.contains('Post de prueba').click()

        cy.contains('button', 'Comment').should('be.disabled')

        cy.get('textarea[placeholder*="comment"]').type('Algo')
        cy.contains('button', 'Comment').should('not.be.disabled')
    })

    it('should navigate back to post list', () => {
        cy.intercept('GET', '**/api/posts/1', {
            statusCode: 200,
            body: {
                id: 1,
                title: 'Post de prueba',
                content: 'Contenido',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }
        })

        cy.intercept('GET', '**/api/posts/1/comments', {
            statusCode: 200,
            body: []
        })

        cy.contains('Post de prueba').click()
        cy.contains('← Back').click()

        cy.contains('Create New Post').should('be.visible')
        cy.contains('Posts').should('be.visible')
    })
})
