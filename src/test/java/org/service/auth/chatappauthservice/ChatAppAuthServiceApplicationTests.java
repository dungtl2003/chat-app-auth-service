package org.service.auth.chatappauthservice;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.service.AuthTokenService;
import org.service.auth.chatappauthservice.service.UserService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.server.LocalServerPort;
import org.springframework.core.io.ClassPathResource;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.result.MockMvcResultMatchers;

import java.io.IOException;
import java.io.InputStream;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@AutoConfigureMockMvc
class ChatAppAuthServiceApplicationTests {

	private static List<User> tempUsers;

	private static String authenticateUrl;

	private static String authorizationUrl;

	@LocalServerPort
	private static int port;

	@Autowired
	private MockMvc mockMvc;

	@Autowired
	private UserService userService;

	@Autowired
	private AuthTokenService authTokenService;

	@BeforeAll
    public static void setup() {
        authenticateUrl = STR."http://localhost:\{port}/api/v1/auth/login";
        authorizationUrl = STR."http://localhost:\{port}/api/v1/auth/authorize";
        tempUsers = getSampleUserFromJson();
    }

	public static JsonNode convertStringToJson(String msg) throws JsonProcessingException {
		ObjectMapper mapper = new ObjectMapper();
		return mapper.readTree(msg);
	}

	private static String buildAuthenticateJsonBodyRequest(User user) throws JsonProcessingException {
		ObjectMapper mapper = JsonMapper.builder().findAndAddModules().build();
		ObjectNode root = mapper.createObjectNode();

		ObjectNode metadata = mapper.createObjectNode();
		metadata.put("method", "authenticate_user");

		ObjectNode payload = mapper.createObjectNode();

		ObjectNode userJson = mapper.createObjectNode();
		userJson.put("email", user.getEmail());
		userJson.put("password", user.getPassword());

		payload.set("user", userJson);

		root.set("metadata", metadata);
		root.set("payload", payload);

		return mapper.writeValueAsString(root);
	}

	private static List<User> getSampleUserFromJson() {
		try (InputStream is = new ClassPathResource("user-sample.json").getInputStream()) {
			ObjectMapper mapper = JsonMapper.builder().findAndAddModules().build();
			return Arrays.asList(mapper.readValue(is, User[].class));
		}
		catch (IOException e) {
			throw new RuntimeException(e);
		}
	}

	private void addTempUsers() {
		tempUsers.forEach(user -> userService.add(user.clone()));
	}

	private void removeTempUsers() {
		tempUsers.forEach(user -> userService.delete(user.getUserId()));
	}

	@Test
	public void testAuthenticateSuccessRequestShouldProvideAccessAndRefreshToken() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			addTempUsers();

			String responseBody = mockMvc
				.perform(MockMvcRequestBuilders.post(authenticateUrl)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isOk())
				.andExpect(MockMvcResultMatchers.header().exists(HttpHeaders.SET_COOKIE))
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode jsonBody = convertStringToJson(responseBody);
			assertNotNull(jsonBody.get("access_token"));
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	// @Test
	// public void testAccessTokenShouldHaveRightFormat() {
	// int size = tempUsers.size();
	// User randomUser = tempUsers.get((int) (Math.random() * size));
	//
	// try {
	// addTempUsers();
	//
	// MvcResult result = mockMvc
	// .perform(MockMvcRequestBuilders.post(authenticateUrl)
	// .contentType(MediaType.APPLICATION_JSON)
	// .content(buildAuthenticateJsonBodyRequest(randomUser)))
	// .andReturn();
	//
	// String accessToken = result.getResponse().getHeader(HttpHeaders.AUTHORIZATION);
	// assertNotNull(accessToken);
	// assertEquals(accessToken.indexOf("Bearer"), 0);
	// }
	// catch (Exception e) {
	// fail(e.getMessage());
	// }
	// finally {
	// removeTempUsers();
	// }
	// }

	@Test
	public void testAuthenticateNonExistedUserRequestShouldGet404NotFound() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			mockMvc
				.perform(MockMvcRequestBuilders.post(authenticateUrl)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isNotFound());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthenticateInvalidUserRequestShouldGet422UnprocessableEntity() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			addTempUsers();

			randomUser.setPassword("fakepassword"); // change password to invalid one

			mockMvc
				.perform(MockMvcRequestBuilders.post(authenticateUrl)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isUnprocessableEntity());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	@Test
	public void testAccessAuthenticateApiWithWrongHttpMethodShouldGet405MethodNotAllowed() {
		try {
			mockMvc.perform(MockMvcRequestBuilders.get(authenticateUrl))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());

			mockMvc.perform(MockMvcRequestBuilders.patch(authenticateUrl))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());

			mockMvc.perform(MockMvcRequestBuilders.put(authenticateUrl))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthorizeWithoutTokenShouldGet401UnauthorizedWithMissingCredentialMessage() {
		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(authorizationUrl))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Missing credential", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthorizeWithWrongTokenFormatShouldGet401UnauthorizedWithInvalidFormatMessage() {
		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(authorizationUrl).header("authorization", "abcd"))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Invalid format", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthorizeWithInvalidTokenShouldGet401UnauthorizedWithInvalidTokenMessage() {
		try {
			String body = mockMvc
				.perform(MockMvcRequestBuilders.get(authorizationUrl).header("authorization", "Bearer faketokenhehehe"))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Invalid token", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
    public void testAuthorizeWithValidTokenShouldGet200WithAuthorizedMessage() {
        String validToken = authTokenService.createAccessToken(tempUsers.getFirst());
        try {
            String body = mockMvc
                    .perform(MockMvcRequestBuilders
                            .get(authorizationUrl)
                            .header("authorization", STR."Bearer \{validToken}"))
                    .andExpect(MockMvcResultMatchers.status().isUnauthorized())
                    .andReturn().getResponse().getContentAsString();

            JsonNode bodyJson = convertStringToJson(body);
            assertEquals("Authorized", bodyJson.get("message").asText());
        } catch (Exception e) {
            fail(e.getMessage());
        }
    }

}
