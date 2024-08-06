package org.service.auth.chatappauthservice.service;

import jakarta.persistence.EntityNotFoundException;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;

import java.math.BigInteger;

public interface UserService {

	void add(User user);

	void deleteUserById(BigInteger id);

	User getValidUser(String email, String password) throws UserNotFoundException, InvalidUserException;

	void addRefreshToken(BigInteger userId, String refreshToken);

	void updateUserRefreshTokens(BigInteger userId, String[] refreshTokens);

	String[] getUserRefreshTokens(BigInteger userId) throws EntityNotFoundException;

}
