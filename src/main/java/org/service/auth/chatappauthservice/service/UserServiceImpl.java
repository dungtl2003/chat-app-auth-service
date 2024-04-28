package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import lombok.NonNull;
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
	public void deleteUserById(long id) {
		userRepository.deleteById(id);
	}

	@Override
	public User getValidUser(@NonNull String email, @NonNull String password)
			throws UserNotFoundException, InvalidUserException {
		User user = getUserByEmail(email);

		if (!isValid(email, password, user)) {
			throw new InvalidUserException("Invalid email and password");
		}

		return user;
	}

	@Override
	public void addRefreshToken(long userId, String refreshToken) throws UserNotFoundException {
		User user = getUserById(userId);

		String[] refreshTokensFromDb = user.getRefreshTokens();
		String[] refreshTokens = refreshTokensFromDb != null
				? Arrays.copyOf(refreshTokensFromDb, refreshTokensFromDb.length + 1) : new String[1];

		refreshTokens[refreshTokens.length - 1] = refreshToken;
		user.setRefreshTokens(refreshTokens);
		userRepository.save(user);
	}

	@Override
	public void updateUserRefreshTokens(long userId, String[] refreshTokens) throws UserNotFoundException {
		User user = getUserById(userId);
		user.setRefreshTokens(refreshTokens);
		userRepository.save(user);
	}

	@Override
	public String[] getUserRefreshTokens(long userId) throws UserNotFoundException {
		return getUserById(userId).getRefreshTokens();
	}

	private User getUserByEmail(String email) throws UserNotFoundException {
        return userRepository
                .findByEmail(email)
                .orElseThrow(() -> new UserNotFoundException(STR."User with email \{email} does not exist"));
    }

	private User getUserById(Long id) throws UserNotFoundException {
        return userRepository
                .findById(id)
                .orElseThrow(() -> new UserNotFoundException(STR."User with id \{id} does not exist"));
    }

	private boolean isValid(String email, String password, User user) {
		return email.equals(user.getEmail()) && passwordEncoder.matches(password, user.getPassword());
	}

}
