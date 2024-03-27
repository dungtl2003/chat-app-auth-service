package org.service.auth.chatappauthservice.service;

import org.service.auth.chatappauthservice.entity.User;

public interface UserService {

	User getValidUser(String email, String password);

	void add(User user);

	void delete(Long userId);

	void updateUserRefreshTokens(User user, String refreshToken);

}
