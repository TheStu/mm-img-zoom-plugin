import axios from 'axios';

export async function fetchData() {
    try {
        const response = await axios.get('http://localhost:8065/plugins/com.tcg.followers/follow');
        return response.data;
    } catch (error) {
        // Handle errors by throwing them
        throw new Error(error.response.data.message || 'An error occurred while fetching the data');
    }
}
