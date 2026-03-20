#pragma once

#ifdef SIMULATION_EXPORTS
#define SIMULATION_API __declspec(dllexport)
#else
#define SIMULATION_API __declspec(dllimport)
#endif

namespace simulation {
class SIMULATION_API Simulation {
public:
    Simulation();
    void add_body(float mass, float x, float y, float z);
    void step(float dt);
    void render();
    int body_count() const;
private:
    struct Impl;
    Impl* impl_;
};
}
