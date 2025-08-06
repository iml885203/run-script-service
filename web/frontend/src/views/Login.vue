<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>🔐 Authentication Required</h1>
        <p>Please enter the secret key to access the dashboard</p>
      </div>
      
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label for="secretKey">Secret Key</label>
          <input
            id="secretKey"
            v-model="secretKey"
            type="password"
            placeholder="Enter your secret key"
            required
            :disabled="isLoading"
            class="form-control"
          />
        </div>
        
        <div v-if="errorMessage" class="error-message">
          {{ errorMessage }}
        </div>
        
        <button 
          type="submit" 
          :disabled="isLoading || !secretKey.trim()"
          class="login-button"
        >
          <span v-if="isLoading">Authenticating...</span>
          <span v-else>Sign In</span>
        </button>
      </form>
      
      <div class="login-footer">
        <p class="help-text">
          If you don't have a secret key, check the service logs for the generated key
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { login, isAuthenticated } = useAuth()

const secretKey = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

const handleLogin = async () => {
  if (!secretKey.value.trim()) {
    errorMessage.value = 'Please enter a secret key'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    const success = await login(secretKey.value)
    
    if (success) {
      // Clear the secret key from memory
      secretKey.value = ''
      
      // Redirect to dashboard
      router.push('/')
    } else {
      errorMessage.value = 'Invalid secret key. Please try again.'
    }
  } catch (error) {
    console.error('Login error:', error)
    errorMessage.value = 'Authentication failed. Please check your connection and try again.'
  } finally {
    isLoading.value = false
  }
}

// Redirect if already authenticated
if (isAuthenticated.value) {
  router.push('/')
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  background: white;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  color: #333;
  margin-bottom: 8px;
  font-size: 24px;
  font-weight: 600;
}

.login-header p {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.login-form {
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #333;
  font-weight: 500;
  font-size: 14px;
}

.form-control {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-control:focus {
  outline: none;
  border-color: #667eea;
}

.form-control:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

.error-message {
  background-color: #fee;
  border: 1px solid #fcc;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 20px;
  color: #c33;
  font-size: 14px;
  text-align: center;
}

.login-button {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.login-button:hover:not(:disabled) {
  opacity: 0.9;
}

.login-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-footer {
  text-align: center;
  margin-top: 20px;
}

.help-text {
  color: #888;
  font-size: 12px;
  line-height: 1.4;
  margin: 0;
}
</style>