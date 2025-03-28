/**
 * Handles HTTP responses by checking the status and returning the parsed JSON.
 * Throws an error if the response is not ok.
 */
const handleResponse = async (response: any) => {
  if (!response.ok) {
    const errorData = await response.json();

    throw new Error(
      errorData.error || errorData.message || "Something went wrong"
    );
  }

  // Parse and return the response JSON
  try {
    return await response.json();
  } catch (err) {
    return;
  }
};

/**
 * Handles HTTP errors in a consistent way.
 */
const handleError = (error: unknown) => {
  console.error("HTTP Error:", error);
  throw error;
};

/**
 * Performs a POST request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - The data to send in the body of the request.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const post = async (endpoint: string, data: any, options?: any) => {
  try {
    const response = await fetch(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
      ...options,
    });
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a GET request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const get = async (endpoint: string, options?: any) => {
  try {
    const response = await fetch(endpoint, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
      ...options,
    });
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a PUT request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - The data to send in the body of the request.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const put = async (endpoint: string, data: any, options?: any) => {
  try {
    const response = await fetch(endpoint, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
      body: JSON.stringify(data),
      ...options,
    });
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a DELETE request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - Signature for the policy deletion.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const remove = async (endpoint: string, data: any, options?: any) => {
  try {
    const response = await fetch(endpoint, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
      ...options,
    });
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};
