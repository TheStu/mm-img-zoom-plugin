import axios from 'axios';

export async function fetchData(siteURL) {
    try {
        const response = await axios.get(`${siteURL}/plugins/com.tcg.followers/follow`);
        return response.data;
    } catch (error) {
        // Handle errors by throwing them
        throw new Error(error.response.data.message || 'An error occurred while fetching the data');
    }
}
