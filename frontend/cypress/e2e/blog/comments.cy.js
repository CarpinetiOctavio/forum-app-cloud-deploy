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

        cy.contains('← Volver').should('be.visible')
        cy.contains('Agregar Comentario').should('be.visible')
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

        cy.get('textarea[placeholder*="comentario"]').type('Mi comentario')
        cy.contains('button', 'Comentar').click()

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

        cy.contains('button', 'Comentar').should('be.disabled')

        cy.get('textarea[placeholder*="comentario"]').type('Algo')
        cy.contains('button', 'Comentar').should('not.be.disabled')
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
        cy.contains('← Volver').click()

        cy.contains('Crear Nuevo Post').should('be.visible')
        cy.contains('Posts').should('be.visible')
    })
})
