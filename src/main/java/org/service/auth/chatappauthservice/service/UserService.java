package org.service.auth.chatappauthservice.service;

import jakarta.persistence.EntityNotFoundException;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;

public interface UserService {

	void add(User user);

	void deleteUserById(String id);

	User getValidUser(String email, String password) throws UserNotFoundException, InvalidUserException;

	void addRefreshToken(String userId, String refreshToken);

	void updateUserRefreshTokens(String userId, String[] refreshTokens);

	String[] getUserRefreshTokens(String userId) throws EntityNotFoundException;

}
