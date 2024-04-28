package org.service.auth.chatappauthservice.service;

import jakarta.persistence.EntityNotFoundException;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;

public interface UserService {

	void add(User user);

	void deleteUserById(long id);

	User getValidUser(String email, String password) throws UserNotFoundException, InvalidUserException;

	void addRefreshToken(long userId, String refreshToken);

	void updateUserRefreshTokens(long userId, String[] refreshTokens);

	String[] getUserRefreshTokens(long userId) throws EntityNotFoundException;

}
