<template>
    <div class="home">
        <div class="header">
            <h1>Welcome to the Home Page</h1>
            <p v-if="username" class="user-info">{{ username }}</p>
        </div>
        <p v-if="pingMessage">Server Response: {{ pingMessage }}</p>
        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </div>
</template>

<script>
import axios from 'axios';

export default {
    name: 'AppHome',
    data() {
        return {
            username: localStorage.getItem('username'),
            pingMessage: '',
            errorMessage: ''
        };
    },
    async mounted() {
        await this.verifyLogin();
    },
    methods: {
        async verifyLogin() {
            try {
                const token = localStorage.getItem('token');
                if (!token) {
                    this.errorMessage = 'Token not found. Please log in.';
                    this.$router.push('/login');
                    return;
                }

                const response = await axios.get('http://localhost:8000/api/v1/ping', {
                    headers: {
                        Authorization: `Bearer ${token}`
                    }
                });

                this.pingMessage = response.data.data || 'Ping successful!';
            } catch (error) {
                console.error('Ping failed:', error);
                this.errorMessage = 'Failed to verify login. Please try again.';
                this.$router.push('/login');
            }
        }
    }
};
</script>

<style scoped>
.header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px;
}

.user-info {
    font-weight: bold;
    font-size: 16px;
    position: absolute;
    top: 10px;
    right: 10px;
}

.error {
    color: red;
}
</style>
