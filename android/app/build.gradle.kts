plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
}

fun config(name: String, default: String = ""): String =
    (project.findProperty("pocketaide.$name") as String?)
        ?: System.getenv(name.replace(Regex("([a-z])([A-Z])"), "$1_$2").uppercase())
        ?: default

fun quoted(value: String) = "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"") + "\""

android {
    namespace = "com.dlddu.pocketaide"
    compileSdk = 36

    defaultConfig {
        applicationId = "com.dlddu.pocketaide"
        minSdk = 26
        targetSdk = 36
        versionCode = config("versionCode", "1").toInt()
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        buildConfigField("String", "FIREBASE_PROJECT_ID", quoted(config("firebaseProjectId", "dlddu-play-console")))
        buildConfigField("String", "FIREBASE_API_KEY", quoted(config("firebaseApiKey")))
        buildConfigField("String", "FIREBASE_SENDER_ID", quoted(config("firebaseSenderId", "657421381831")))
    }

    signingConfigs {
        create("release") {
            val store = config("signingStoreFile")
            if (store.isNotEmpty()) {
                storeFile = file(store)
                storePassword = config("signingStorePassword")
                keyAlias = config("signingKeyAlias")
                keyPassword = config("signingKeyPassword")
            }
        }
    }

    buildTypes {
        debug {
            applicationIdSuffix = ".debug"
            buildConfigField("String", "BACKEND_BASE_URL", quoted(config("backendBaseUrl", "http://10.0.2.2:8080")))
            buildConfigField("String", "FIREBASE_APP_ID", quoted(config("firebaseAppId", "1:657421381831:android:16491d777af54099e84ee2")))
            // mock-exception: EXT — 실 GitHub 은 전용 테스트 계정 PAT 가 CI 시크릿에 없어 E2E 가 부를 수 없다; 디버그 빌드만 GITHUB_API_BASE_URL 이 있을 때 앱의 GitHubClient 를 로컬 GitHub API 스텁으로 향하게 한다 (docs/e2e-mocking-policy.md)
            buildConfigField("String", "GITHUB_API_BASE_URL", quoted(config("githubApiBaseUrl", "https://api.github.com")))
        }
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            signingConfig = signingConfigs.getByName("release").takeIf { it.storeFile != null }
            buildConfigField("String", "BACKEND_BASE_URL", quoted(config("backendBaseUrl")))
            buildConfigField("String", "FIREBASE_APP_ID", quoted(config("firebaseAppId", "1:657421381831:android:c98abab847b61adbe84ee2")))
            buildConfigField("String", "GITHUB_API_BASE_URL", quoted("https://api.github.com"))
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    testOptions {
        unitTests.isReturnDefaultValues = true
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.fragment.ktx)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.process)
    implementation(libs.androidx.browser)
    implementation(libs.androidx.glance.appwidget)
    implementation(libs.androidx.glance.material3)
    implementation(libs.androidx.work.runtime.ktx)

    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.tooling.preview)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons.extended)
    debugImplementation(libs.compose.ui.tooling)

    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.okhttp)

    implementation(platform(libs.firebase.bom))
    implementation(libs.firebase.messaging)

    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.okhttp.mockwebserver)

    androidTestImplementation(platform(libs.compose.bom))
    androidTestImplementation(libs.compose.ui.test.junit4)
    androidTestImplementation(libs.androidx.test.ext.junit)
    androidTestImplementation(libs.androidx.test.runner)
    androidTestImplementation(libs.androidx.test.rules)
    androidTestImplementation(libs.androidx.test.uiautomator)
    androidTestImplementation(libs.androidx.test.espresso.core)
    androidTestImplementation(libs.okhttp)
    androidTestImplementation(libs.kotlinx.serialization.json)
    debugImplementation(libs.compose.ui.test.manifest)
}
