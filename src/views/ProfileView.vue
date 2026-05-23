<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import * as zod from 'zod'
import { useAuth } from '@/composables/useAuth'
import { getProfile, updateProfile, type UpdateProfilePayload } from '@/api/auth'

const router = useRouter()
const { user } = useAuth()
const isSubmitting = ref(false)

// Валидация
const profileSchema = zod.object({
  name: zod.string().min(2, 'Минимум 2 символа'),
  surname: zod.string().optional(),
  department: zod.string().optional(),
  password: zod.string().min(8, 'Минимум 8 символов').optional().or(zod.literal('')),
  passwordConfirm: zod.string().optional().or(zod.literal('')),
}).refine(data => !data.password || data.password === data.passwordConfirm, {
  message: 'Пароли не совпадают',
  path: ['passwordConfirm'],
})

const { values: form, handleSubmit, setValues, setFieldError } = useForm<UpdateProfilePayload>({
  initialValues: { name: '', surname: '', department: '', password: '', passwordConfirm: '' }
})

// Загрузка данных
onMounted(async () => {
  const res = await getProfile()
  if (res?.data) {
    setValues({
      name: res.data.name ?? '',
      surname: res.data.surname ?? '',
      department: res.data.department ?? '',
      password: '',
      passwordConfirm: ''
    })
    console.log('✅ [Profile] Form populated')
  } else {
    console.warn('⚠️ [Profile] No data returned from getProfile()')
  }
})

// Сабмит
const onSubmit = handleSubmit(async (values) => {
  const result = profileSchema.safeParse(values)
  if (!result.success) {
    result.error.issues.forEach(issue => {
      const field = issue.path[0] as keyof UpdateProfilePayload
      if (field) setFieldError(field, issue.message)
    })
    return
  }

  isSubmitting.value = true
  try {
    const res = await updateProfile(result.data)
    if (res.data) {
      router.push({ name: 'Dashboard' })
    }
  } catch (e: any) {
    setFieldError('name', e?.message || 'Ошибка сохранения')
  } finally {
    isSubmitting.value = false
  }
})
</script>

<template>
  <!-- 🚫 НЕТ <AppLayout>! Vue Router сам вставит этот блок внутрь лейаута -->
  <div class="max-w-2xl mx-auto p-6 bg-white rounded-xl shadow-sm">
    <h1 class="text-2xl font-bold mb-6 text-gray-800">Мой профиль</h1>
    
    <form @submit.prevent="onSubmit" class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">Имя *</label>
          <input v-model="form.name" type="text" class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
          <p class="text-red-500 text-xs mt-1">{{ form.name ? '' : '' }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Фамилия</label>
          <input v-model="form.surname" type="text" class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
        </div>
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Логин</label>
        <input :value="user?.login" disabled class="w-full px-3 py-2 border rounded-lg bg-gray-100 text-gray-500 cursor-not-allowed" />
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Почта</label>
        <input :value="user?.email" disabled class="w-full px-3 py-2 border rounded-lg bg-gray-100 text-gray-500 cursor-not-allowed" />
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Департамент</label>
        <input v-model="form.department" type="text" class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
      </div>

      <div class="border-t pt-4 mt-4">
        <h2 class="text-lg font-semibold mb-3">Смена пароля</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">Новый пароль</label>
            <input v-model="form.password" type="password" class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Подтверждение</label>
            <input v-model="form.passwordConfirm" type="password" class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
          </div>
        </div>
      </div>

      <div class="flex gap-3 pt-2">
        <button type="submit" :disabled="isSubmitting" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition">
          {{ isSubmitting ? 'Сохранение...' : 'Сохранить' }}
        </button>
        <button type="button" @click="router.back()" class="px-4 py-2 border rounded-lg hover:bg-gray-50 transition">
          Отмена
        </button>
      </div>
    </form>
  </div>
</template>