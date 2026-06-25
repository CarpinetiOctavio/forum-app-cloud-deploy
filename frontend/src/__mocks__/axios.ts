// Full axios mock for tests
const axiosMock = {
    get: jest.fn(() => Promise.resolve({ data: {} })),
    post: jest.fn(() => Promise.resolve({ data: {} })),
    put: jest.fn(() => Promise.resolve({ data: {} })),
    delete: jest.fn(() => Promise.resolve({ data: {} })),
    create: jest.fn(),
  };

  // Configure create to return the same mock
  axiosMock.create.mockReturnValue(axiosMock);

  export default axiosMock;
