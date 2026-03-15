// SentenceViewModel.swift
// PocketAide

import Foundation
import SwiftUI

/// 문장 모음 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class SentenceViewModel: ObservableObject {

    // MARK: - Published State

    @Published var categories: [SentenceCategory] = []
    @Published var sentences: [Sentence] = []
    @Published var isLoading = false
    @Published var errorMessage: String? = nil
    @Published var showAddCategorySheet = false
    @Published var showAddSentenceSheet = false
    @Published var editingSentence: Sentence? = nil

    // MARK: - Dependencies

    private let sentenceService = SentenceService()
    private let keychainService = KeychainService()

    // MARK: - Init

    init() {}

    // MARK: - Computed

    /// 특정 카테고리에 속하는 문장을 반환합니다.
    func sentences(for category: SentenceCategory) -> [Sentence] {
        sentences.filter { $0.categoryId == category.id }
    }

    // MARK: - Public Methods

    /// 카테고리와 문장 목록을 서버에서 불러옵니다.
    func loadData() async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            async let fetchedCategories = sentenceService.listCategories(serverURL: serverURL, token: token)
            async let fetchedSentences = sentenceService.listSentences(serverURL: serverURL, token: token)
            categories = try await fetchedCategories
            sentences = try await fetchedSentences
        } catch {
            errorMessage = "데이터를 불러오지 못했습니다."
        }
    }

    /// 새 카테고리를 생성합니다.
    func createCategory(name: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await sentenceService.createCategory(
                name: name,
                serverURL: serverURL,
                token: token
            )
            categories.append(created)
        } catch {
            errorMessage = "카테고리를 생성하지 못했습니다."
        }
    }

    /// 새 문장을 생성합니다.
    func createSentence(content: String, categoryId: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await sentenceService.createSentence(
                content: content,
                categoryId: categoryId,
                serverURL: serverURL,
                token: token
            )
            sentences.append(created)
        } catch {
            errorMessage = "문장을 생성하지 못했습니다."
        }
    }

    /// 문장을 수정합니다.
    func updateSentence(id: Int, content: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let updated = try await sentenceService.updateSentence(
                id: id,
                content: content,
                serverURL: serverURL,
                token: token
            )
            if let idx = sentences.firstIndex(where: { $0.id == id }) {
                sentences[idx] = updated
            }
        } catch {
            errorMessage = "문장을 수정하지 못했습니다."
        }
    }

    /// 문장을 삭제합니다.
    func deleteSentence(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await sentenceService.deleteSentence(id: id, serverURL: serverURL, token: token)
            sentences.removeAll { $0.id == id }
        } catch {
            errorMessage = "문장을 삭제하지 못했습니다."
        }
    }
}
