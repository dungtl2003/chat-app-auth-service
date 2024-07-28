package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import lombok.NonNull;
import org.service.auth.chatappauthservice.constants.StatusMessage;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.repository.UserRepository;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.Arrays;

@AllArgsConstructor
@Service
public class UserServiceImpl implements UserService {

	private final UserRepository userRepository;

	private final PasswordEncoder passwordEncoder;

	@Override
	public void add(User user) {
		user.setPassword(passwordEncoder.encode(user.getPassword()));
		userRepository.save(user);
	}

	@Override
	public void deleteUserById(String id) {
		userRepository.deleteById(id);
	}

	@Override
	public User getValidUser(@NonNull String email, @NonNull String password)
			throws UserNotFoundException, InvalidUserException {
		User user = getUserByEmail(email);

		if (!isValid(email, password, user)) {
			throw new InvalidUserException(StatusMessage.INVALID_EMAIL_PASSWORD);
		}

		return user;
	}

	@Override
	public void addRefreshToken(String userId, String refreshToken) throws UserNotFoundException {
		User user = getUserById(userId);

		String[] refreshTokensFromDb = user.getRefreshTokens();
		String[] refreshTokens = Arrays.copyOf(refreshTokensFromDb, refreshTokensFromDb.length + 1);

		refreshTokens[refreshTokens.length - 1] = refreshToken;
		user.setRefreshTokens(refreshTokens);
		userRepository.save(user);
	}

	@Override
	public void updateUserRefreshTokens(String userId, String[] refreshTokens) throws UserNotFoundException {
		User user = getUserById(userId);
		user.setRefreshTokens(refreshTokens);
		userRepository.save(user);
	}

	@Override
	public String[] getUserRefreshTokens(String userId) throws UserNotFoundException {
		return getUserById(userId).getRefreshTokens();
	}

	private User getUserByEmail(String email) throws UserNotFoundException {
		return userRepository.findByEmail(email)
			.orElseThrow(() -> new UserNotFoundException(StatusMessage.EMAIL_NOT_FOUND));
	}

	private User getUserById(String id) throws UserNotFoundException {
		return userRepository.findById(id).orElseThrow(() -> new UserNotFoundException(StatusMessage.ID_NOT_FOUND));
	}

	private boolean isValid(String email, String password, User user) {
		return email.equals(user.getEmail()) && passwordEncoder.matches(password, user.getPassword());
	}

}
