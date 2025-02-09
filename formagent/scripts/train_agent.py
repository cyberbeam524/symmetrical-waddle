from stable_baselines3 import PPO
from scripts.form_filling_env import FormFillingEnv

env = FormFillingEnv()
model = PPO("MlpPolicy", env, verbose=1)
model.learn(total_timesteps=10000)
model.save("../models/rl_agent")
