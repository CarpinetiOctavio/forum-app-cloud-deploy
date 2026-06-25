describe('Posts Management', () => {
    beforeEach(() => {
        // Log in first
        cy.visit('/')

        cy.intercept('POST', '**/api/auth/login', {
            statusCode: 200,
            body: {
                id: 1,
                email: 'test@example.com',
                username: 'testuser'
            }
        })

        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: []
        }).as('getPosts')

        cy.get('input#email').type('test@example.com')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.wait('@getPosts')
    })

    it('should show message when there are no posts', () => {
        cy.contains('No posts yet').should('be.visible')
    })

    it('should create a post successfully', () => {
        cy.intercept('POST', '**/api/posts', {
            statusCode: 201,
            body: {
                id: 1,
                title: 'Mi primer post',
                content: 'Contenido de prueba',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }
        }).as('createPost')

        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: [{
                id: 1,
                title: 'Mi primer post',
                content: 'Contenido de prueba',
                user_id: 1,
                username: 'testuser',
                created_at: new Date().toISOString()
            }]
        })

        cy.get('input[placeholder*="title"]').type('Mi primer post')
        cy.get('textarea[placeholder*="share"]').type('Contenido de prueba')
        cy.contains('button', 'Publish Post').click()

        cy.wait('@createPost')

        // Verify the post appears in the list
        cy.contains('Mi primer post').should('be.visible')
        cy.contains('Contenido de prueba').should('be.visible')
    })

    it('should show error when creating post without title', () => {
        cy.get('textarea[placeholder*="share"]').type('Solo contenido')
        cy.contains('button', 'Publish Post').click()

        // HTML5 validation prevents submit
        cy.get('input[placeholder*="title"]').should('have.prop', 'validity')
            .and('have.property', 'valueMissing', true)
    })

    it('should list existing posts', () => {
        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: [
                {
                    id: 1,
                    title: 'Post 1',
                    content: 'Contenido 1',
                    user_id: 1,
                    username: 'testuser',
                    created_at: '2024-10-27T00:00:00Z'
                },
                {
                    id: 2,
                    title: 'Post 2',
                    content: 'Contenido 2',
                    user_id: 2,
                    username: 'otheruser',
                    created_at: '2024-10-27T00:00:00Z'
                }
            ]
        })

        cy.visit('/')
        cy.get('input#email').type('test@example.com')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.contains('Post 1').should('be.visible')
        cy.contains('Post 2').should('be.visible')
        cy.contains('@testuser').should('be.visible')
        cy.contains('@otheruser').should('be.visible')
    })

    it("should not show delete button on other users' posts", () => {
        cy.intercept('GET', '**/api/posts', {
            statusCode: 200,
            body: [{
                id: 2,
                title: 'Post de otro',
                content: 'Contenido',
                user_id: 2,
                username: 'otheruser',
                created_at: new Date().toISOString()
            }]
        })

        cy.visit('/')
        cy.get('input#email').type('test@example.com')
        cy.get('input#password').type('123456')
        cy.get('button[type="submit"]').click()

        cy.contains('Post de otro').should('be.visible')
        cy.get('.post-card').within(() => {
            cy.contains('Eliminar').should('not.exist')
        })
    })
})
