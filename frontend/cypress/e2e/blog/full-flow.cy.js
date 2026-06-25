describe('Full User Flow', () => {
    it('full flow: register → create post → comment → logout', () => {
        cy.visit('/')

        // 1. REGISTRATION
        cy.intercept('POST', '**/api/auth/register', {
            statusCode: 201,
            body: { id: 1, email: 'nuevo@example.com', username: 'nuevo' }
        }).as('register')

        cy.intercept('GET', '**/api/posts', { statusCode: 200, body: [] })

        cy.contains('¿No tienes cuenta? Regístrate').click()
        cy.get('input#email').type('nuevo@example.com')
        cy.get('input#username').type('nuevo')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.wait('@register')
        cy.contains('Hola, @nuevo').should('be.visible')

        // 2. CREATE POST
        cy.intercept('POST', '**/api/posts', {
            statusCode: 201,
            body: {
                id: 1,
                title: 'Mi primer post',
                content: 'Contenido inicial',
                user_id: 1,
                username: 'nuevo',
                created_at: new Date().toISOString()
            }
        }).as('createPost')

        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: [{
                id: 1,
                title: 'Mi primer post',
                content: 'Contenido inicial',
                user_id: 1,
                username: 'nuevo',
                created_at: new Date().toISOString()
            }]
        })

        cy.get('input[placeholder*="título"]').type('Mi primer post')
        cy.get('textarea[placeholder*="compartir"]').type('Contenido inicial')
        cy.contains('button', 'Publicar Post').click()

        cy.wait('@createPost')
        cy.contains('Mi primer post').should('be.visible')

        // 3. VIEW DETAIL AND COMMENT
        cy.intercept('GET', '**/api/posts/1', {
            statusCode: 200,
            body: {
                id: 1,
                title: 'Mi primer post',
                content: 'Contenido inicial',
                user_id: 1,
                username: 'nuevo',
                created_at: new Date().toISOString()
            }
        })

        cy.intercept('GET', '**/api/posts/1/comments', {
            statusCode: 200,
            body: []
        })

        cy.intercept('POST', '**/api/posts/1/comments', {
            statusCode: 201,
            body: {
                id: 1,
                post_id: 1,
                user_id: 1,
                username: 'nuevo',
                content: 'Gran post!',
                created_at: new Date().toISOString()
            }
        })

        cy.contains('Mi primer post').click()
        cy.get('textarea[placeholder*="comentario"]').type('Gran post!')
        cy.contains('button', 'Comentar').click()

        // 4. LOGOUT
        cy.contains('← Volver').click()
        cy.contains('Cerrar Sesión').click()

        cy.get('h2').should('contain', 'Iniciar Sesión')
    })
})
