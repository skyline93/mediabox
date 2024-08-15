<template>
    <div class="library-container">
        <h2>Library Page</h2>
        <button @click="handleIndex">创建索引</button>
    </div>
</template>

<script setup>
import axios from 'axios';
import { ref } from 'vue';
import { BASE_URL } from '../config';

const errorMessage = ref('');

const token = localStorage.getItem('token');

const handleIndex = async () => {
    try {
        const response = await axios.post(
            `${BASE_URL}/api/v1/photo/import`,
            {},
            {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            }
        );
        console.log('Import successful:', response.data);
    } catch (error) {
        console.error('Import failed:', error);
        errorMessage.value = 'Import failed. Please try again.';
    }
};
</script>

<style scoped>
.library-container {
    max-width: 600px;
    margin: 20px auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 5px;
    box-shadow: 2px 2px 12px rgba(0, 0, 0, 0.1);
}

button {
    padding: 10px 20px;
    background-color: #42b983;
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
}

button:hover {
    background-color: #369b72;
}
</style>