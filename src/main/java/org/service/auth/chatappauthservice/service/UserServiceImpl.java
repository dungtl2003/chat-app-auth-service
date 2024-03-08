package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import lombok.NonNull;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.repository.UserRepository;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@Service
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;

    private final UserDTOMapper userDTOMapper;

    private final PasswordEncoder passwordEncoder;

    @Override
    public UserDTO getValidUser(@NonNull String email, @NonNull String password) throws UserNotFoundException, InvalidUserException {
        UserDTO user = getUserByEmail(email);

        if (!isValid(email, password, user)) {
            throw new InvalidUserException("Invalid email and password");
        }

        return user;
    }

    @Override
    public void add(@NonNull User user) {
        user.setPassword(passwordEncoder.encode(user.getPassword()));
        userRepository.save(user);
    }

    @Override
    public void delete(@NonNull Long userId) {
        userRepository.deleteById(userId);
    }

    private UserDTO getUserByEmail(String email) throws UserNotFoundException {
        return userRepository
                .findByEmail(email)
                .map(userDTOMapper)
                .orElseThrow(() -> new UserNotFoundException(STR."User with email \{email} does not exist"));
    }

    private boolean isValid(String email, String password, UserDTO user) {
        return email.equals(user.email()) && passwordEncoder.matches(password, user.password());
    }
}
