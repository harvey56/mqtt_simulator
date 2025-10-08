import React, { useState, useRef } from "react";
import {
  Upload,
  File,
  ChevronRight,
  ChevronDown,
  Edit3,
  Save,
  X,
  Plus,
  FolderOpen,
  Settings,
  Trash2,
  FolderPlus,
  Play,
  Square,
} from "lucide-react";

interface JsonNode {
  key: string;
  type: string;
  value: any;
  children?: JsonNode[];
  path: string;
  isExpanded?: boolean;
  isEditing?: boolean;
  originalValue?: any;
  isSelected?: boolean;
}

interface Configuration {
  id: string;
  name: string;
  topic: { type: string; value: string };
  frequency: { type: string; value: string };
  jsonData: JsonNode[];
  fileName: string;
}

interface Project {
  id: string;
  name: string;
  isRunning?: boolean;
  configurations: Configuration[];
}

interface JsonDashboardProps {}

const JsonDashboard: React.FC<JsonDashboardProps> = () => {
  const [projects, setProjects] = useState<Project[]>([
    {
      id: "1",
      name: "Sample Project",
      configurations: [],
    },
  ]);
  const [selectedProject, setSelectedProject] = useState<string>("1");
  const [selectedConfiguration, setSelectedConfiguration] = useState<
    string | null
  >(null);
  const [currentConfig, setCurrentConfig] = useState<Configuration>({
    id: "",
    name: "",
    topic: { type: "string", value: "" },
    frequency: { type: "string", value: "" },
    jsonData: [],
    fileName: "",
  });
  const [isDragOver, setIsDragOver] = useState(false);
  const [showNewProjectForm, setShowNewProjectForm] = useState(false);
  const [newProjectName, setNewProjectName] = useState("");
  const [runningProjects, setRunningProjects] = useState<Set<string>>(
    new Set()
  );
  const fileInputRef = useRef<HTMLInputElement>(null);

  const getValueType = (value: any): string => {
    if (value === null) return "null";
    if (Array.isArray(value)) return "array";
    if (typeof value === "object") return "object";
    return typeof value;
  };

  const parseJsonToTree = (obj: any, parentPath = ""): JsonNode[] => {
    const nodes: JsonNode[] = [];

    Object.entries(obj).forEach(([key, value]) => {
      const path = parentPath ? `${parentPath}.${key}` : key;
      const type = getValueType(value);

      const node: JsonNode = {
        key,
        type,
        value,
        path,
        isExpanded: true,
        isEditing: false,
        originalValue: value,
        isSelected: false,
      };

      if (type === "object" && value !== null) {
        node.children = parseJsonToTree(value, path);
      } else if (type === "array") {
        node.children = (value as any[]).map((item, index) => ({
          key: `[${index}]`,
          type: getValueType(item),
          value: item,
          path: `${path}[${index}]`,
          isExpanded: true,
          isEditing: false,
          originalValue: item,
          isSelected: false,
          children:
            getValueType(item) === "object" && item !== null
              ? parseJsonToTree(item, `${path}[${index}]`)
              : undefined,
        }));
      }

      nodes.push(node);
    });

    return nodes;
  };

  const handleFileSelect = (file: File) => {
    if (file.type === "application/json" || file.name.endsWith(".json")) {
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const jsonContent = JSON.parse(e.target?.result as string);
          const treeData = parseJsonToTree(jsonContent);
          setCurrentConfig((prev) => ({
            ...prev,
            jsonData: treeData,
            fileName: file.name,
            name: file.name.replace(".json", ""),
          }));
        } catch (error) {
          alert("Invalid JSON file. Please select a valid JSON file.");
        }
      };
      reader.readAsText(file);
    } else {
      alert("Please select a JSON file.");
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    const files = e.dataTransfer.files;
    if (files[0]) {
      handleFileSelect(files[0]);
    }
  };

  const createProject = () => {
    if (!newProjectName.trim()) {
      alert("Please enter a project name");
      return;
    }

    const newProject: Project = {
      id: Date.now().toString(),
      name: newProjectName.trim(),
      configurations: [],
    };

    setProjects((prev) => [...prev, newProject]);
    setSelectedProject(newProject.id);
    setNewProjectName("");
    setShowNewProjectForm(false);
  };

  const deleteProject = (projectId: string) => {
    if (projects.length <= 1) {
      alert("Cannot delete the last project");
      return;
    }

    if (
      confirm(
        "Are you sure you want to delete this project and all its configurations?"
      )
    ) {
      setProjects((prev) => prev.filter((p) => p.id !== projectId));

      if (selectedProject === projectId) {
        const remainingProjects = projects.filter((p) => p.id !== projectId);
        setSelectedProject(remainingProjects[0].id);
        setCurrentConfig({
          id: "",
          name: "",
          topic: { type: "string", value: "" },
          frequency: { type: "string", value: "" },
          jsonData: [],
          fileName: "",
        });
        setSelectedConfiguration(null);
      }
    }
  };

  const startProject = (projectId: string) => {
    const project = projects.find((p) => p.id === projectId);
    if (!project || project.configurations.length === 0) {
      alert("Cannot start project: No configurations found");
      return;
    }

    setRunningProjects((prev) => new Set([...prev, projectId]));

    // Simulate API request start
    console.log(`Starting API requests for project: ${project.name}`);
    console.log("Configurations:", project.configurations);

    // Here you would implement actual API request logic
    alert(`Started API requests for project: ${project.name}`);
  };

  const stopProject = (projectId: string) => {
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;

    setRunningProjects((prev) => {
      const newSet = new Set(prev);
      newSet.delete(projectId);
      return newSet;
    });

    // Simulate API request stop
    console.log(`Stopping API requests for project: ${project.name}`);

    // Here you would implement actual API request stop logic
    alert(`Stopped API requests for project: ${project.name}`);
  };

  const toggleExpanded = (path: string) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, isExpanded: !node.isExpanded };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const toggleEdit = (path: string) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, isEditing: !node.isEditing };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const toggleSelection = (path: string) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, isSelected: !node.isSelected };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const updateValue = (path: string, newValue: any) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, value: newValue };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const saveEdit = (path: string) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, isEditing: false, originalValue: node.value };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const cancelEdit = (path: string) => {
    const updateNodes = (nodes: JsonNode[]): JsonNode[] => {
      return nodes.map((node) => {
        if (node.path === path) {
          return { ...node, isEditing: false, value: node.originalValue };
        }
        if (node.children) {
          return { ...node, children: updateNodes(node.children) };
        }
        return node;
      });
    };
    setCurrentConfig((prev) => ({
      ...prev,
      jsonData: updateNodes(prev.jsonData),
    }));
  };

  const saveConfiguration = () => {
    if (!currentConfig.name.trim()) {
      alert("Please provide a configuration name");
      return;
    }

    const configId = selectedConfiguration || Date.now().toString();
    const updatedConfig = { ...currentConfig, id: configId };

    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === selectedProject) {
          const existingConfigIndex = project.configurations.findIndex(
            (c) => c.id === configId
          );
          if (existingConfigIndex >= 0) {
            // Update existing configuration
            const updatedConfigurations = [...project.configurations];
            updatedConfigurations[existingConfigIndex] = updatedConfig;
            return { ...project, configurations: updatedConfigurations };
          } else {
            // Add new configuration
            return {
              ...project,
              configurations: [...project.configurations, updatedConfig],
            };
          }
        }
        return project;
      })
    );

    setSelectedConfiguration(configId);
    alert("Configuration saved successfully!");
  };

  const addToProject = () => {
    if (!currentConfig.name.trim() || currentConfig.jsonData.length === 0) {
      alert("Please provide a configuration name and upload a JSON file");
      return;
    }

    const newConfigId = Date.now().toString();
    const newConfig = { ...currentConfig, id: newConfigId };

    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === selectedProject) {
          return {
            ...project,
            configurations: [...project.configurations, newConfig],
          };
        }
        return project;
      })
    );

    // Reset current config
    const resetConfig = {
      id: "",
      name: "",
      topic: { type: "string", value: "" },
      frequency: { type: "string", value: "" },
      jsonData: [],
      fileName: "",
    };
    setCurrentConfig({
      ...resetConfig,
    });
    setSelectedConfiguration(null);

    // Clear file input
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }

    alert("Configuration added to project successfully!");
  };

  const loadConfiguration = (config: Configuration) => {
    setCurrentConfig(config);
    setSelectedConfiguration(config.id);
  };

  const deleteConfiguration = (configId: string) => {
    if (confirm("Are you sure you want to delete this configuration?")) {
      setProjects((prev) =>
        prev.map((project) => {
          if (project.id === selectedProject) {
            return {
              ...project,
              configurations: project.configurations.filter(
                (c) => c.id !== configId
              ),
            };
          }
          return project;
        })
      );

      if (selectedConfiguration === configId) {
        setCurrentConfig({
          id: "",
          name: "",
          topic: { type: "string", value: "" },
          frequency: { type: "string", value: "" },
          jsonData: [],
          fileName: "",
        });
        setSelectedConfiguration(null);
      }
    }
  };

  const renderJsonNode = (node: JsonNode, depth = 0) => {
    const hasChildren = node.children && node.children.length > 0;
    const indent = depth * 24;

    return (
      <React.Fragment key={node.path}>
        <div className="group hover:bg-gray-50 transition-colors duration-200">
          <div className="grid grid-cols-12 gap-4 py-3 px-4 border-b border-gray-100">
            <div
              className="col-span-4 flex items-center"
              style={{ paddingLeft: `${indent}px` }}
            >
              {hasChildren && (
                <button
                  onClick={() => toggleExpanded(node.path)}
                  className="mr-2 p-1 hover:bg-gray-200 rounded transition-colors duration-200"
                >
                  {node.isExpanded ? (
                    <ChevronDown className="w-4 h-4 text-gray-600" />
                  ) : (
                    <ChevronRight className="w-4 h-4 text-gray-600" />
                  )}
                </button>
              )}
              <span className="font-medium text-gray-800 truncate">
                {node.key}
              </span>
            </div>
            <div className="col-span-2 flex items-center">
              <span
                className={`px-2 py-1 rounded-full text-xs font-medium ${
                  node.type === "string"
                    ? "bg-green-100 text-green-700"
                    : node.type === "number"
                    ? "bg-blue-100 text-blue-700"
                    : node.type === "boolean"
                    ? "bg-purple-100 text-purple-700"
                    : node.type === "object"
                    ? "bg-orange-100 text-orange-700"
                    : node.type === "array"
                    ? "bg-red-100 text-red-700"
                    : "bg-gray-100 text-gray-700"
                }`}
              >
                {node.type}
              </span>
            </div>
            <div className="col-span-5 flex items-center">
              {node.type === "object" || node.type === "array" ? (
                <span className="text-gray-500 italic">
                  {node.type === "array"
                    ? `Array[${node.children?.length || 0}]`
                    : "Object"}
                </span>
              ) : (
                <div className="flex items-center space-x-2 w-full">
                  {node.isEditing ? (
                    <div className="flex items-center space-x-2 w-full">
                      <input
                        type={node.type === "number" ? "number" : "text"}
                        value={node.value}
                        onChange={(e) => updateValue(node.path, e.target.value)}
                        className="flex-1 px-3 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                      <button
                        onClick={() => saveEdit(node.path)}
                        className="p-1 text-green-600 hover:bg-green-100 rounded transition-colors duration-200"
                      >
                        <Save className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => cancelEdit(node.path)}
                        className="p-1 text-red-600 hover:bg-red-100 rounded transition-colors duration-200"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ) : (
                    <div className="flex items-center space-x-2 w-full">
                      <span className="flex-1 text-gray-700 truncate">
                        {node.type === "string"
                          ? `"${node.value}"`
                          : String(node.value)}
                      </span>
                      <button
                        onClick={() => toggleEdit(node.path)}
                        className="opacity-0 group-hover:opacity-100 p-1 text-blue-600 hover:bg-blue-100 rounded transition-all duration-200"
                      >
                        <Edit3 className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
            <div className="col-span-1 flex items-center justify-center">
              <input
                type="checkbox"
                checked={node.isSelected || false}
                onChange={() => toggleSelection(node.path)}
                className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 focus:ring-2"
              />
            </div>
          </div>
        </div>
        {hasChildren &&
          node.isExpanded &&
          node.children?.map((child) => renderJsonNode(child, depth + 1))}
      </React.Fragment>
    );
  };

  return (
    <div className="w-full bg-gradient-to-br from-blue-50 to-indigo-100 flex">
      {/* Left Sidebar - Projects */}
      <div className="w-96 bg-white shadow-xl border-r border-gray-200 flex flex-col">
        <div className="bg-gradient-to-r from-indigo-600 to-purple-600 px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-white flex items-center">
              <FolderOpen className="w-5 h-5 mr-2" />
              Projects
            </h2>
            <button
              onClick={() => setShowNewProjectForm(true)}
              className="p-2 text-white hover:bg-white/20 rounded-lg transition-colors duration-200"
              title="Create New Project"
            >
              <FolderPlus className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* New Project Form */}
        {showNewProjectForm && (
          <div className="p-4 bg-blue-50 border-b border-blue-200">
            <div className="space-y-3">
              <input
                type="text"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="Enter project name"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                onKeyDown={(e) => e.key === "Enter" && createProject()}
              />
              <div className="flex space-x-2">
                <button
                  onClick={createProject}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-md font-medium transition-colors duration-200"
                >
                  Create
                </button>
                <button
                  onClick={() => {
                    setShowNewProjectForm(false);
                    setNewProjectName("");
                  }}
                  className="flex-1 bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-2 rounded-md font-medium transition-colors duration-200"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4">
          <div className="space-y-4">
            {projects.map((project) => (
              <div
                key={project.id}
                className="border border-gray-200 rounded-lg overflow-hidden"
              >
                <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <h3 className="font-semibold text-gray-800 flex items-center">
                        <FolderOpen className="w-4 h-4 mr-2 text-gray-600" />
                        {project.name}
                      </h3>
                      {runningProjects.has(project.id) && (
                        <span className="px-2 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-full">
                          Running
                        </span>
                      )}
                    </div>
                    <div className="flex items-center space-x-1">
                      {!runningProjects.has(project.id) ? (
                        <button
                          onClick={() => startProject(project.id)}
                          className="p-1 text-green-600 hover:bg-green-100 rounded transition-colors duration-200"
                          title="Start API Requests"
                          disabled={project.configurations.length === 0}
                        >
                          <Play className="w-4 h-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => stopProject(project.id)}
                          className="p-1 text-red-600 hover:bg-red-100 rounded transition-colors duration-200"
                          title="Stop API Requests"
                        >
                          <Square className="w-4 h-4" />
                        </button>
                      )}
                      <button
                        onClick={() => deleteProject(project.id)}
                        className="p-1 text-red-500 hover:bg-red-100 rounded transition-colors duration-200"
                        title="Delete Project"
                        disabled={runningProjects.has(project.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
                <div className="p-3">
                  {project.configurations.length === 0 ? (
                    <p className="text-gray-500 text-sm italic">
                      No configurations yet
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {project.configurations.map((config) => (
                        <div
                          key={config.id}
                          className={`p-3 rounded-lg border cursor-pointer transition-all duration-200 ${
                            selectedConfiguration === config.id
                              ? "bg-blue-50 border-blue-300 shadow-sm"
                              : "bg-white border-gray-200 hover:bg-gray-50"
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div
                              onClick={() => loadConfiguration(config)}
                              className="flex-1"
                            >
                              <div className="flex items-center space-x-2">
                                <Settings className="w-4 h-4 text-gray-600" />
                                <span className="font-medium text-gray-700">
                                  {config.name}
                                </span>
                              </div>
                              <p className="text-xs text-gray-500 mt-1">
                                {config.fileName}
                              </p>
                            </div>
                            <button
                              onClick={() => deleteConfiguration(config.id)}
                              className="p-1 text-red-500 hover:bg-red-100 rounded transition-colors duration-200"
                              title="Delete Configuration"
                              disabled={runningProjects.has(project.id)}
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 p-6">
        <div className="max-w-full mx-auto">
          <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
            {/* Header */}
            <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-8 py-6">
              <h1 className="text-3xl font-bold text-white mb-2">
                JSON Dashboard
              </h1>
              <p className="text-blue-100">
                Upload and edit JSON files with an intuitive tree interface
              </p>
            </div>

            {/* Configuration Settings */}
            <div className="p-6 border-b border-gray-200 bg-gray-50">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Configuration Name
                  </label>
                  <input
                    type="text"
                    value={currentConfig.name}
                    onChange={(e) =>
                      setCurrentConfig((prev) => ({
                        ...prev,
                        name: e.target.value,
                      }))
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder="Enter configuration name"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Project
                  </label>
                  <select
                    value={selectedProject}
                    onChange={(e) => setSelectedProject(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    {projects.map((project) => (
                      <option key={project.id} value={project.id}>
                        {project.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Topic
                  </label>
                  <div className="flex space-x-2">
                    <select
                      value={currentConfig.topic.type}
                      onChange={(e) =>
                        setCurrentConfig((prev) => ({
                          ...prev,
                          topic: { ...prev.topic, type: e.target.value },
                        }))
                      }
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="string">String</option>
                      <option value="number">Number</option>
                      <option value="boolean">Boolean</option>
                    </select>
                    <input
                      type="text"
                      value={currentConfig.topic.value}
                      onChange={(e) =>
                        setCurrentConfig((prev) => ({
                          ...prev,
                          topic: { ...prev.topic, value: e.target.value },
                        }))
                      }
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Topic value"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Frequency
                  </label>
                  <div className="flex space-x-2">
                    <select
                      value={currentConfig.frequency.type}
                      onChange={(e) =>
                        setCurrentConfig((prev) => ({
                          ...prev,
                          frequency: {
                            ...prev.frequency,
                            type: e.target.value,
                          },
                        }))
                      }
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="string">String</option>
                      <option value="number">Number</option>
                      <option value="boolean">Boolean</option>
                    </select>
                    <input
                      type="text"
                      value={currentConfig.frequency.value}
                      onChange={(e) =>
                        setCurrentConfig((prev) => ({
                          ...prev,
                          frequency: {
                            ...prev.frequency,
                            value: e.target.value,
                          },
                        }))
                      }
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Frequency value"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Upload Section */}
            <div className="p-8 border-b border-gray-200">
              <div
                className={`relative border-2 border-dashed rounded-xl p-8 transition-all duration-300 ${
                  isDragOver
                    ? "border-blue-400 bg-blue-50"
                    : "border-gray-300 hover:border-blue-400 hover:bg-gray-50"
                }`}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
              >
                <div className="text-center">
                  <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-lg font-medium text-gray-700 mb-2">
                    Drop your JSON file here or click to browse
                  </p>
                  <p className="text-sm text-gray-500 mb-4">
                    Supports .json files up to 10MB
                  </p>
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium transition-colors duration-200"
                  >
                    Select File
                  </button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".json"
                    onChange={handleFileUpload}
                    className="hidden"
                  />
                </div>
              </div>

              {currentConfig.fileName && (
                <div className="mt-4 flex items-center justify-between">
                  <div className="flex items-center space-x-2 text-sm text-gray-600">
                    <File className="w-4 h-4" />
                    <span>Loaded: {currentConfig.fileName}</span>
                  </div>
                  <div className="flex space-x-3">
                    <button
                      onClick={saveConfiguration}
                      className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center space-x-2"
                      disabled={!currentConfig.name.trim()}
                    >
                      <Save className="w-4 h-4" />
                      <span>Save Configuration</span>
                    </button>
                    <button
                      onClick={addToProject}
                      className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center space-x-2 disabled:bg-gray-400 disabled:cursor-not-allowed"
                      disabled={
                        !currentConfig.name.trim() ||
                        currentConfig.jsonData.length === 0
                      }
                    >
                      <Plus className="w-4 h-4" />
                      <span>Add to Project</span>
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Data Table */}
            {currentConfig.jsonData.length > 0 && (
              <div className="overflow-hidden">
                <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                  <div className="grid grid-cols-12 gap-4 text-sm font-semibold text-gray-700 uppercase tracking-wide">
                    <div className="col-span-4">Key</div>
                    <div className="col-span-2">Type</div>
                    <div className="col-span-5">Value</div>
                    <div className="col-span-1">Random</div>
                  </div>
                </div>
                <div className="max-h-96 overflow-y-auto">
                  {currentConfig.jsonData.map((node) => renderJsonNode(node))}
                </div>
              </div>
            )}

            {currentConfig.jsonData.length === 0 && (
              <div className="text-center py-12">
                <div className="text-gray-400 mb-4">
                  <File className="w-16 h-16 mx-auto mb-4 opacity-50" />
                  <p className="text-lg font-medium">No JSON file loaded</p>
                  <p className="text-sm">Upload a JSON file to start editing</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default JsonDashboard;
