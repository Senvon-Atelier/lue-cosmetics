import axios from 'axios';

// Configure axios defaults globally (used by Orval-generated functions)
axios.defaults.baseURL = import.meta.env.VITE_API_URL || '/api/v1';
axios.defaults.withCredentials = true; // Send cookies for session auth
axios.defaults.headers.common['Content-Type'] = 'application/json';

// Response interceptor for error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle common error cases
    if (error.response) {
      // Server responded with error status
      const errorData = error.response.data;
      if (errorData?.error) {
        throw new Error(errorData.error.message || 'An error occurred');
      }
      throw new Error(errorData?.message || `Request failed with status ${error.response.status}`);
    } else if (error.request) {
      // Request was made but no response received
      throw new Error('No response from server. Please check your connection.');
    } else {
      // Something else happened
      throw new Error(error.message || 'An unexpected error occurred');
    }
  }
);

// Export configured axios instance for direct use if needed
export const apiClient = axios;
